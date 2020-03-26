package testserver

import (
	"google.golang.org/grpc"
	pb_test "grpcTestProject/pb"
	"net"
	"strconv"
)

func RegisterServer(s *grpc.Server)  {
	pb_test.RegisterMyTestServer(s, &TestServer{})
}
const serverName = "mytest"
func ServerUp()  {
	port, err := GetFreePort()
	if err != nil {
		panic(err)
	}

	localIp, err := GetLocalIP()
	if err != nil {
		panic(err)
	}

	conn, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	RegisterServer(s)
	err = RegisterServerToConsul(serverName, localIp, strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	if err := s.Serve(conn); err != nil {
		panic(err)
	}
}
