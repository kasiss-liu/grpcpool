package main

import (
	"context"
	"fmt"
	"net"

	"github.com/kasiss-liu/grpcpool/_examples/helloworld/protos/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type helloServer struct{}

func (h *helloServer) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloResp, error) {
	return &helloworld.HelloResp{Words: "hello " + req.Name}, nil
}

func main() {
	listen, err := net.Listen("tcp", ":8889")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	server := grpc.NewServer()
	helloworld.RegisterHelloWorldServer(server, &helloServer{})
	reflection.Register(server)

	err = server.Serve(listen)
	if err != nil {
		fmt.Println(err)
	}
}
