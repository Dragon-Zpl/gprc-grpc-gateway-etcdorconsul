package testserver

import (
	"fmt"
	"golang.org/x/net/context"
	"grpcTestProject/pb"
)

type TestServer struct {

}

func (*TestServer) HelloWord(ctx context.Context, in *pb_test.Empty) (*pb_test.TestResp, error) {
	fmt.Println("come in")
	return &pb_test.TestResp{
		Word:                 "hello world!",
	}, nil
}

