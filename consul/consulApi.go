package consul

import (
	"encoding/json"
	consulapi "github.com/hashicorp/consul/api"
)
var (
	ConsulClient *consulapi.Client
)
const (
	consulUrl ="192.168.189.30:8500"
)


func init() {
	var err error
	config := consulapi.DefaultConfig()
	config.Address = consulUrl
	ConsulClient, err = consulapi.NewClient(config)
	if err != nil {
		panic(err)
	}
}




type ConsulKVData struct {
	Name string
	Labels []string
	Type string
}

func GetKeyData(key string) []byte {
	kv, _, err := ConsulClient.KV().Get(key, nil)
	if err != nil || kv == nil {
		return nil
	}
	return kv.Value
}

func SetKeyValue(key string, value interface{}) error {
	data, _ := json.Marshal(value)
	_, err := ConsulClient.KV().Put(&consulapi.KVPair{
		Key:         key,
		Value:       data,
	}, nil)
	return err
}

func UpdateValue(key string, value []ConsulKVData) error {
	data, _ := json.Marshal(value)
	_, err := ConsulClient.KV().Put(&consulapi.KVPair{
		Key:         key,
		Value:       data,
	}, nil)
	return err
}

func DeleteData(key string) error {
	_, err := ConsulClient.KV().Delete(key, nil)
	if err != nil {
		return err
	}
	return nil
}

func GetAllKey(prefix string) []string {
	allKeys, _, _ := ConsulClient.KV().Keys(prefix, "", nil)
	return allKeys
}

