package testserver


import (
	"context"
	"encoding/json"
	"fmt"
	etcd3 "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"google.golang.org/grpc/grpclog"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

//var RegistryDir = conf.GetEtcdConf().Dir

var etcd3Client *etcd3.Client

type Registrar struct {
	etcd3Client *etcd3.Client
	key         string
	value       string
	ttl         time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
}

type Option struct {
	EtcdConfig     etcd3.Config
	RegistryDir    string
	ServiceName    string
	ServiceVersion string
	NodeID         string
	NData          NodeData
	Ttl            time.Duration
}

//func NewOption(etcdConf etcd3.Config, serviceName string, serviceVersion string, ttl time.Duration, metadata ...map[string]string) *Option {
//	// todo 获取node id
//	return &Option{
//		EtcdConfig:     etcdConf,
//		RegistryDir:    RegistryDir,
//		ServiceName:    serviceName,
//		ServiceVersion: "",
//		NodeID:         "",
//		NData: NodeData{
//			// 获取本机ip
//			Addr:     "",
//			Metadata: metadata[0],
//		},
//		Ttl: ttl,
//	}
//}

type NodeData struct {
	Addr     string
	Metadata map[string]string
}

func NewRegistrar(option Option) (*Registrar, error) {
	if etcd3Client == nil {
		var err error
		etcd3Client, err = etcd3.New(option.EtcdConfig)
		if err != nil {
			return nil, err
		}

	}

	val, err := json.Marshal(option.NData)
	if err != nil {
		return nil, err
	}

	CreateNodeId(etcd3Client, &option)

	ctx, cancel := context.WithCancel(context.Background())
	registry := &Registrar{
		etcd3Client: etcd3Client,
		key:         option.RegistryDir + "/" + option.ServiceName + "/" + option.ServiceVersion + "/" + option.NodeID,
		value:       string(val),
		ttl:         option.Ttl,
		ctx:         ctx,
		cancel:      cancel,
	}
	return registry, nil
}

func NewTtlEtcdKey(config etcd3.Config, key, val string, ttl time.Duration) (*Registrar, error) {
	if etcd3Client == nil {
		var err error
		etcd3Client, err = etcd3.New(config)
		if err != nil {
			return nil, err
		}

	}

	ctx, cancel := context.WithCancel(context.Background())
	registry := &Registrar{
		etcd3Client: etcd3Client,
		key:         key,
		value:       val,
		ttl:         ttl,
		ctx:         ctx,
		cancel:      cancel,
	}
	return registry, nil
}

func CreateNodeId(client *etcd3.Client, option *Option) error {
	resp, err := client.KV.Get(context.Background(), option.RegistryDir+"/"+option.ServiceName+"/"+option.ServiceVersion+"/", etcd3.WithPrefix(), etcd3.WithKeysOnly())
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		option.NodeID = "1"
		return nil
	}
	ids := make(map[string]struct{})
	for _, v := range resp.Kvs {
		splits := strings.Split(string(v.Key), "/")
		nodeId := splits[len(splits)-1]
		ids[nodeId] = struct{}{}
	}
	for {
		nodeId := strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(resp.Kvs) + 1))
		if _, ok := ids[nodeId]; !ok {
			option.NodeID = nodeId
			return nil
		}
	}
}

func (e *Registrar) Register() error {

	insertFunc := func() error {
		resp, err := e.etcd3Client.Grant(e.ctx, int64(e.ttl))
		if err != nil {
			fmt.Printf("[Register] %v\n", err.Error())
			return err
		}
		_, err = e.etcd3Client.Get(e.ctx, e.key)
		if err != nil {
			if err == rpctypes.ErrKeyNotFound {
				if _, err := e.etcd3Client.Put(e.ctx, e.key, e.value, etcd3.WithLease(resp.ID)); err != nil {
					grpclog.Infof("grpclb: set key '%s' with ttl to etcd3 failed: %s", e.key, err.Error())
				}
			} else {
				grpclog.Infof("grpclb: key '%s' connect to etcd3 failed: %s", e.key, err.Error())
			}
			return err
		} else {
			// refresh set to true for not notifying the watcher
			if _, err := e.etcd3Client.Put(e.ctx, e.key, e.value, etcd3.WithLease(resp.ID)); err != nil {
				grpclog.Infof("grpclb: refresh key '%s' with ttl to etcd3 failed: %s", e.key, err.Error())
				return err
			}
		}
		return nil
	}

	err := insertFunc()
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second * (e.ttl / 5))
	for {
		select {
		case <-ticker.C:
			insertFunc()
		case <-e.ctx.Done():
			ticker.Stop()
			if _, err := e.etcd3Client.Delete(context.Background(), e.key); err != nil {
				grpclog.Infof("grpclb: deregister '%s' failed: %s", e.key, err.Error())
			}
			return nil
		}
	}

	return nil
}

func (e *Registrar) Unregister() error {
	e.cancel()
	return nil
}
