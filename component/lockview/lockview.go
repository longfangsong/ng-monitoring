package lockview

import (
	"sync"
	"time"

	"github.com/genjidb/genji"
	"github.com/pingcap/log"
	"github.com/pingcap/ng-monitoring/component/topology"
	"github.com/pingcap/ng-monitoring/config/pdvariable"
)

var (
	store   Store
	manager Manager
)

type Manager struct {
	sync.Mutex
	currentListening []topology.Component
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
		for {
			time.Sleep(30 * time.Second)
			manager.Lock()
			currentListening := manager.currentListening
			manager.Unlock()
			for _, tidbComponent := range currentListening {
				client := tipb.
			}
		}
	}()
	return nil
}
