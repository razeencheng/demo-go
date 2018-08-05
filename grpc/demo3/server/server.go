package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"google.golang.org/grpc/credentials"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"

	pb "github.com/razeencheng/demo-go/grpc/demo3/helloworld"
)

type SayHelloServer struct{}

// 服务器端流式 RPC, 接收一次客户端请求，返回一个流
func (s *SayHelloServer) ListHello(in *pb.HelloWorldRequest, stream pb.HelloWorldService_ListHelloServer) error {
	log.Printf("Client Say: %v", in.Greeting)

	stream.Send(&pb.HelloWorldResponse{Reply: "ListHello Reply " + in.Greeting + " 1"})
	time.Sleep(1 * time.Second)
	stream.Send(&pb.HelloWorldResponse{Reply: "ListHello Reply " + in.Greeting + " 2"})
	time.Sleep(1 * time.Second)
	stream.Send(&pb.HelloWorldResponse{Reply: "ListHello Reply " + in.Greeting + " 3"})
	time.Sleep(1 * time.Second)
	return nil
}

// 客户端流式 RPC， 客户端流式请求，服务器可返回一次
func (s *SayHelloServer) SayMoreHello(stream pb.HelloWorldService_SayMoreHelloServer) error {
	// 接受客户端请求
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		log.Printf("SayMoreHello Client Say: %v", req.Greeting)
	}

	// 流读取完成后，返回
	return stream.SendAndClose(&pb.HelloWorldResponse{Reply: "SayMoreHello Recv Muti Greeting"})
}

func (s *SayHelloServer) SayHelloChat(stream pb.HelloWorldService_SayHelloChatServer) error {

	go func() {
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				return
			}

			log.Printf("SayHelloChat Client Say: %v", req.Greeting)
		}
	}()

	stream.Send(&pb.HelloWorldRequest{Greeting: "SayHelloChat Server Say Hello 1"})
	time.Sleep(1 * time.Second)
	stream.Send(&pb.HelloWorldRequest{Greeting: "SayHelloChat Server Say Hello 2"})
	time.Sleep(1 * time.Second)
	stream.Send(&pb.HelloWorldRequest{Greeting: "SayHelloChat Server Say Hello 3"})
	time.Sleep(1 * time.Second)
	return nil
}

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

	// 加载证书和密钥 （同时能验证证书与私钥是否匹配）
	cert, err := tls.LoadX509KeyPair("certs/test_server.pem", "certs/test_server.key")
	if err != nil {
		panic(err)
	}

	// 将根证书加入证书词
	// 测试证书的根如果不加入可信池，那么测试证书将视为不可惜，无法通过验证。
	certPool := x509.NewCertPool()
	rootBuf, err := ioutil.ReadFile("certs/root.pem")
	if err != nil {
		panic(err)
	}

	if !certPool.AppendCertsFromPEM(rootBuf) {
		panic("fail to append test ca")
	}

	tlsConf := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    certPool,
	}

	serverOpt := grpc.Creds(credentials.NewTLS(tlsConf))
	grpcServer := grpc.NewServer(serverOpt)

	pb.RegisterHelloWorldServiceServer(grpcServer, &SayHelloServer{})

	log.Println("Server Start...")
	grpcServer.Serve(lis)
}
