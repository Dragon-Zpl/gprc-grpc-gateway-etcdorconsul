package testserver

import (
	"google.golang.org/grpc"
	pb_test "grpcTestProject/pb"
)
func ClientUp() pb_test.MyTestClient {
	c, err := grpc.Dial("localhost:8090", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return pb_test.NewMyTestClient(c)
}

