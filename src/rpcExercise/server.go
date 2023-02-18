package main

import (
	"log"
	"net"
	"net/rpc"
)

//其中Hello方法必须满足Go语言的RPC规则：
//1 方法只能有两个可序列化的参数
//2 其中第二个参数是指针类型，
//3 并且返回一个error类型，
//4 同时必须是公开的方法。

type HelloService struct{}

func (h *HelloService) Hello(request string, reply *string) error {
	*reply = "hello:" + request
	return nil
}

func main() {
	//然后就可以将HelloService类型的对象注册为一个RPC服务：
	//其中rpc.Register函数调用会将对象类型中所有满足RPC规则的对象方法注册为RPC函数，
	//所有注册的方法会放在“HelloService”服务空间之下

	err := rpc.RegisterName("HelloService", new(HelloService))
	if err != nil {
		panic(err)
	}
	//然后我们建立一个唯一的TCP链接，并且通过rpc.ServeConn函数在该TCP链接上为对方提供RPC服务。
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal("Accept error:", err)
	}

	rpc.ServeConn(conn)
}
