package meta

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetaJson(t *testing.T) {
	target := ProfileTarget{
		Kind:      "profile",
		Component: "tidb",
		Address:   "10.0.1.21:10080",
	}
	data, err := json.Marshal(target)
	require.NoError(t, err)
	require.Equal(t, `{"kind":"profile","component":"tidb","address":"10.0.1.21:10080"}`, string(data))

	param := BasicQueryParam{
		Begin:      1,
		End:        2,
		Limit:      100,
		Targets:    []ProfileTarget{target},
		DataFormat: ProfileDataFormatProtobuf,
	}
	data, err = json.Marshal(param)
	require.NoError(t, err)
	require.Equal(t, `{"begin_time":1,"end_time":2,"limit":100,"targets":[{"kind":"profile","component":"tidb","address":"10.0.1.21:10080"}],"data_format":"protobuf"}`, string(data))

	list := ProfileList{
		Target: target,
		TsList: []int64{1, 2, 3, 4},
	}
	data, err = json.Marshal(list)
	require.NoError(t, err)
	require.Equal(t, `{"target":{"kind":"profile","component":"tidb","address":"10.0.1.21:10080"},"timestamp_list":[1,2,3,4]}`, string(data))
}
