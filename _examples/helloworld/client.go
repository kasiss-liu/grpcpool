package main

import (
	"context"
	"fmt"
	"os"
	"time"

	pool "github.com/kasiss-liu/grpcpool"
	"github.com/kasiss-liu/grpcpool/_examples/helloworld/protos/helloworld"
	"google.golang.org/grpc"
)

//这里要warp一层 创建client的方法
func newServerClient(cc grpc.ClientConnInterface) interface{} {
	ccc := cc.(*grpc.ClientConn)
	return helloworld.NewHelloWorldClient(ccc)
}

//一个构造器
func buildServiceCenter() *pool.ServiceCenter {
	scb := pool.NewServiceCenterBuilder()

	builders := map[string]pool.ServerBuilderFunc{
		"say": newServerClient,
	}

	err := scb.SetServerWithDefaultOptions("hello", builders, "127.0.0.1:8889")
	if err != nil {
		return nil
	}
	return scb.Build()
}

var sc *pool.ServiceCenter = buildServiceCenter()

func main() {
	resp, err := SayHello()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	fmt.Println(resp)
}

//SayHello to call server function
func SayHello() (response string, err error) {
	/* 使用安全方法获取一个server 并发起请求
	c,ok := sc.Get("hello")
	if !ok {
		return "",errors.New("unexpected server setting")
	}
	client,release,err := c.GetServerClient("say")
	if err != nil {
		return
	}
	defer release()
	*/

	//在上层确保服务存在的情况下 不再验证server存在 可使用UnsafeGet方法获取
	client, release, err := sc.UnsafeGet("hello").GetServerClient("say")
	if err != nil {
		return
	}
	defer release()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var resp *helloworld.HelloResp
	req := &helloworld.HelloRequest{Name: "Jason"}
	resp, err = client.(helloworld.HelloWorldClient).SayHello(ctx, req)
	if err != nil {
		return
	}
	fmt.Println(resp)
	return
}
