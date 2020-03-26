package consul

import "encoding/json"

const (
	consulDir = "grpc/"
)

type ConsulServerData struct {
	Ip string `json:"ip"`
	Port string `json:"port"`
}

func GetConsulDirData(serverName string) []*ConsulServerData {
	keys := GetAllKey(consulDir + serverName)
	res := make([]*ConsulServerData, 0)
	for _, key := range keys {
		var data ConsulServerData
		err := json.Unmarshal(GetKeyData(key), &data)
		if err == nil {
			res = append(res, &data)
		}
	}

	return res
}