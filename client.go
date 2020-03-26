package main

import (
	"context"
	"fmt"
	pb_test "grpcTestProject/pb"
	"grpcTestProject/testserver"
)

func main() {
	client := testserver.ClientUp()
	res, err := client.HelloWord(context.Background(), &pb_test.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Word)
}