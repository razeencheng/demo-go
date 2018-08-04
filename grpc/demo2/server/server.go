package main

import (
	"context"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"

	pb "github.com/razeencheng/demo-go/grpc/demo2/helloworld"
)

type SayHelloServer struct{}

func (s *SayHelloServer) SayHelloWorld(ctx context.Context, in *pb.HelloWorldRequest) (res *pb.HelloWorldResponse, err error) {
	log.Printf("Client Greeting:%s", in.Greeting)
	log.Printf("Client Info:%v", in.Infos)

	var an *any.Any
	if in.Infos["hello"] == "world" {
		an, err = ptypes.MarshalAny(&pb.HelloWorld{Msg: "Good Request"})
	} else {
		an, err = ptypes.MarshalAny(&pb.Error{Msg: []string{"Bad Request", "Wrong Info Msg"}})
	}

	if err != nil {
		return
	}
	return &pb.HelloWorldResponse{
		Reply:   "Hello World !!",
		Details: []*any.Any{an},
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterHelloWorldServiceServer(grpcServer, &SayHelloServer{})
	grpcServer.Serve(lis)
}
