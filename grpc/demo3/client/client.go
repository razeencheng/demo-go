package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/razeencheng/demo-go/grpc/demo3/helloworld"
)

func main() {
	cert, err := tls.LoadX509KeyPair("certs/test_client.pem", "certs/test_client.key")
	if err != nil {
		panic(err)
	}

	// 将根证书加入证书池
	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile("certs/root.pem")
	if err != nil {
		panic(err)
	}

	if !certPool.AppendCertsFromPEM(bs) {
		panic("cc")
	}

	// 新建凭证
	transportCreds := credentials.NewTLS(&tls.Config{
		ServerName:   "server.razeen.me",
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	})

	dialOpt := grpc.WithTransportCredentials(transportCreds)

	conn, err := grpc.Dial("localhost:8080", dialOpt)
	if err != nil {
		log.Fatalf("Dial failed:%v", err)
	}
	defer conn.Close()

	client := pb.NewHelloWorldServiceClient(conn)
	resp1, err := client.SayHelloWorld(context.Background(), &pb.HelloWorldRequest{
		Greeting: "Hello Server 1 !!",
		Infos:    map[string]string{"hello": "world"},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Resp1:%+v", resp1)

	resp2, err := client.SayHelloWorld(context.Background(), &pb.HelloWorldRequest{
		Greeting: "Hello Server 2 !!",
	})
	if err != nil {
		log.Fatalf("%v", err)
	}
	log.Printf("Resp2:%+v", resp2)

	// 服务器端流式 RPC;
	recvListHello, err := client.ListHello(context.Background(), &pb.HelloWorldRequest{Greeting: "Hello Server List Hello"})
	if err != nil {
		log.Fatalf("ListHello err: %v", err)
	}

	for {
		resp, err := recvListHello.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("ListHello Server Resp: %v", resp.Reply)
	}

	// 客户端流式 RPC;
	sayMoreClient, err := client.SayMoreHello(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		sayMoreClient.Send(&pb.HelloWorldRequest{Greeting: fmt.Sprintf("SayMoreHello Hello Server %d", i)})
	}

	sayMoreResp, err := sayMoreClient.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SayMoreHello Server Resp: %v", sayMoreResp.Reply)

	// 双向流式 RPC;
	sayHelloChat, err := client.SayHelloChat(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for i := 0; i < 3; i++ {
			sayHelloChat.Send(&pb.HelloWorldRequest{Greeting: fmt.Sprintf("SayHelloChat Hello Server %d", i)})
		}
	}()

	for {
		resp, err := sayHelloChat.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("SayHelloChat Server Say: %v", resp.Greeting)
	}

}
