package lockview

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/genjidb/genji"
	"github.com/pingcap/log"
	"github.com/pingcap/ng-monitoring/component/topology"
	"github.com/pingcap/ng-monitoring/config/pdvariable"
	"github.com/pingcap/tipb/go-tipb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	store   Store
	manager Manager
)

type Manager struct {
	sync.Mutex
	currentListening []topology.Component
}

func dial(ctx context.Context, tlsConfig *tls.Config, addr string) (*grpc.ClientConn, error) {
	var tlsOption grpc.DialOption
	if tlsConfig == nil {
		tlsOption = grpc.WithInsecure()
	} else {
		tlsOption = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return grpc.DialContext(
		dialCtx,
		addr,
		tlsOption,
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    60 * time.Second,
			Timeout: 10 * time.Second,
		}),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  100 * time.Millisecond, // Default was 1s.
				Multiplier: 1.6,                    // Default
				Jitter:     0.2,                    // Default
				MaxDelay:   30 * time.Second,       // Default was 120s.
			},
		}),
	)
}

func Init(
	gj *genji.DB,
	topSub topology.Subscriber,
	varSub pdvariable.Subscriber,
) (err error) {
	store = NewStore(gj)
	go func() {
		for {
			coms := <-topSub
			if len(coms) == 0 {
				log.Warn("got empty scrapers. Seems to be encountering network problems")
				continue
			}
			currentListening := []topology.Component{}
			for _, com := range coms {
				if com.Name == topology.ComponentTiDB {
					currentListening = append(currentListening, com)
				}
			}
			manager.Lock()
			manager.currentListening = currentListening
			manager.Unlock()
		}
	}()
	go func() {
		// for {
		time.Sleep(30 * time.Second)
		manager.Lock()
		currentListening := manager.currentListening
		manager.Unlock()
		for _, tidbComponent := range currentListening {
			addr := fmt.Sprintf("%s:%d", tidbComponent.IP, tidbComponent.StatusPort)
			log.Info("should connecting to", zap.String("addr", addr))
			conn, err := dial(context.Background(), nil, addr)
			if err != nil {
				log.Error("Error:", zap.Error(err))
			}
			log.Info("connected")
			client := tipb.NewLockViewAgentClient(conn)
			c, err := client.PullTrxRecords(context.Background())
			if err != nil {
				panic(err)
			}
			log.Info("got client")
			received := make(chan *tipb.TrxRecords)
			go func() {
				for {
					r, err := c.Recv()
					if err != nil {
						panic(err)
					}
					received <- r
				}
			}()
			go func() {
				for {
					select {
					case records := <-received:
						for _, record := range records.GetRecords() {
							log.Info("recieved", zap.Uint64("r", record.GetId()))
							store.Insert(record.GetId(), record.GetSqlDigests(), 0)
						}
					case <-time.After(30 * time.Second):
						err := c.Send(&emptypb.Empty{})
						if err != nil {
							panic(err)
						}
					}
				}
			}()
		}
		// }
	}()
	return nil
}
