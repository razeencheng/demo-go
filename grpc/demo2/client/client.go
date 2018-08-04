package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "github.com/razeencheng/demo-go/grpc/demo2/helloworld"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Dial failed:%v", err)
	}
	defer conn.Close()

	client := pb.NewHelloWorldServiceClient(conn)
	resp1, err := client.SayHelloWorld(context.Background(), &pb.HelloWorldRequest{
		Greeting: "Hello Server 1 !!",
		Infos:    map[string]string{"hello": "world"},
	})

	log.Printf("Resp1:%+v", resp1)

	resp2, err := client.SayHelloWorld(context.Background(), &pb.HelloWorldRequest{
		Greeting: "Hello Server 2 !!",
	})

	log.Printf("Resp2:%+v", resp2)
}
