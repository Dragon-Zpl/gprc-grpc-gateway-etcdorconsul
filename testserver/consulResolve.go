package testserver

import (
	"errors"
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
	"grpcTestProject/consul"
	"net"
	"strconv"
	"sync"
	"time"
)



type consulResolver struct {
	consulClient *consulapi.Client
	dir string
	serverName string
	cc            resolver.ClientConn
	wg            sync.WaitGroup
	closeCh       chan bool
}

func (c *consulResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	c.consulClient = consul.ConsulClient
	c.cc = cc
	c.closeCh = make(chan bool)
	c.Watcher()
	return c, nil
}

func (cb *consulResolver) Scheme() string {
	return "consul"
}

//ResolverNow方法什么也不做，因为和consul保持了发布订阅的关系
//不需要像dns_resolver那个定时的去刷新
func (cr *consulResolver) ResolveNow(opt resolver.ResolveNowOption) {
}

//暂时先什么也不做吧
func (cr *consulResolver) Close() {
	cr.closeCh<-true
}

// 监听，没使用consul里面的服务注册只使用到了key/value模仿,如使用服务的话可使用cc.consulClient.Health().severice获取所有注册的服务信息，并检查健康状态
func (cr *consulResolver) Watcher() {
	cr.wg.Add(1)
	go func() {
		defer cr.wg.Done()
		t := time.NewTimer(10 * time.Second)
		for   {
			select {
			case <- t.C:
				datas := consul.GetConsulDirData(cr.serverName)
				address := make([]resolver.Address, 0)
				for _, data := range datas {
					address = append(address, resolver.Address{
						Addr:       data.Ip + ":" + data.Port,
					})
				}
				cr.cc.UpdateState(resolver.State{Addresses:address})
				t.Reset(10 * time.Second)
			case <- cr.closeCh:
				return
			}
		}
	}()
}

func RegisterConsul(serverName string)  {
	resolver.Register(&consulResolver{
		serverName: serverName,
	})
}

const (
	consulDir = "grpc/"
)
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("dont get local ip")
}

func GetFreePort() (int, error) {
	for {
		if port, err := GetFreeFunc(); err == nil{
			return port, nil
		} else if err != nil {
			return port, err
		}
	}
}

// 获取一个空闲可用的端口号
func GetFreeFunc() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}


func RegisterServerToConsul(serverName, ip, port string) error {
	key := consulDir + serverName
	existKey := consul.GetAllKey(key)
	lenKey := len(existKey)
	key += "_" + strconv.Itoa(lenKey + 1)
	data := consul.ConsulServerData{
		Ip:   ip,
		Port: port,
	}
	return consul.SetKeyValue(key, data)
}