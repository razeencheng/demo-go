# [gRPC在Go中的使用（二）gRPC实现简单通讯](https://razeen.me/post/how-to-use-grpc-in-golang-02.html)

Desc:gRPC实现简单通讯,Google 开源 RPC 框架 gRPC 初探



在上一篇中，我们用protobuf定义了两个消息`HelloWorldRequest`与`HelloWorldResponse`以及一个`HelloWorldService`服务。同时，我们还生成了相应的go代码`.pb.go`。

那么客户端与服务端怎么去通过这些接口去完成通讯呢？下面我们一起实现一个简单的gRPC通讯。



<!--more-->

在RPC通讯中，客户端使用存根(SayHelloWorld)发送请求到服务器并且等待响应返回，整个过程就像我们平常函数调用一样。

```rpc
service HelloWorldService {
  rpc SayHelloWorld(HelloWorldRequest) returns (HelloWorldResponse){}
}
```

那么接下来，我们先创建一个服务端。



### 创建服务端


在生成的`hello_world.pb.go`中，已经为我们生成了服务端的接口：

```go
// HelloWorldServiceServer is the server API for HelloWorldService service.
type HelloWorldServiceServer interface {
	SayHelloWorld(context.Context, *HelloWorldRequest) (*HelloWorldResponse, error)
}
```

在服务端我们首先要做的就是实现这个接口。

```go
package main

import (
	"context"
	"log"
	"net"

	pb "github.com/razeencheng/demo-go/grpc/demo2/helloworld"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
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
```

 简单如上面的几行，实现了这个接口我们只需要创建一个结构`SayHelloServer`,同时实现`HelloWorldServiceServer`的所有方法即可。

这里为了演示效果我打印了一些数据，同时利用`any.Any`在不同的情况下返回不同的类型数据。



当然，只是现实了接口还不够，我们还需要启动一个服务，这样客户端才能使用该服务。启动服务很简单，就像我们平常启用一个Server一样。

```go
func main() {
	// 我们首先须监听一个tcp端口
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	
    // 新建一个grpc服务器
	grpcServer := grpc.NewServer()
    // 向grpc服务器注册SayHelloServer
	pb.RegisterHelloWorldServiceServer(grpcServer, &SayHelloServer{})
    // 启动服务
	grpcServer.Serve(lis)
}
```

从上面的代码，我们可以看到，简单的4步即可启动一个服务。

1. 监听一个服务端口，供客户端调用；
2. 创建一个grpc服务器，当然这里可以设置`授权认证`,这个在下一篇中我们将详细介绍；
3. 注册服务，其实是调用生存的`.pb.go`中的`RegisterHelloWorldServiceServer`方法，将我们这里实现的`SayHelloServer`加入到该服务中。
4. 启动服务，等待客户端连接。

我们` go run server.go`,无任何报错，这样一个简单的grpc服务的服务端就准备就绪了。接下来我们看看客户端。



### 创建客户端


例如：

```go
package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "github.com/razeencheng/demo-go/grpc/demo2/helloworld"
)

func main() {
    // 创建一个 gRPC channel 和服务器交互
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Dial failed:%v", err)
	}
	defer conn.Close()

    // 创建客户端
	client := pb.NewHelloWorldServiceClient(conn)
    
    // 直接调用
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
```

客户端的实现比服务端更简洁，三步即可。

1. 创建一个 gRPC channel 和服务器交互。这里也是可以设置`授权认证`的；
2. 创建一个客户端去执行RPC。用到的也是`.pb.go`内的`NewHelloWorldServiceClient`方法；
3. 像函数调用一样去调用RPC服务。



我直接RUN起来，如下，我们可以看到客户端发送到服务的消息以及服务端对不同消息的不同回复。

![](https://st.razeen.me/essay/image/grpc/grpc-result.png)



那么到这里，我们简单的实现了一个gRPC通讯。但很多时候，我们可能希望客户端与服务器能更安全的通信，或者客户端与服务器不再是一种固定的结构的传输，需要流式的去处理一些问题等等。针对这些问题，在下一篇博客中，我将结合实例详细说明。



*文中完整代码在[这里](https://github.com/razeencheng/demo-go/tree/master/grpc/demo2)。*