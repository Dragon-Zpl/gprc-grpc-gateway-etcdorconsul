package testserver

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	pb_test "grpcTestProject/pb"
	"net/http"
)

func GateWayUp()  {
	gwmux := runtime.NewServeMux()
	ctx := context.Background()
	RegisterConsul(serverName)
	err := pb_test.RegisterMyTestHandlerFromEndpoint(ctx, gwmux, "consul:///", []grpc.DialOption{grpc.WithInsecure(), grpc.WithBalancerName(RoundRobin)})
	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(":8070", gwmux)
	if err != nil {
		panic(err)
	}

}