
# [gRPC在Go中的使用（三）gRPC实现TLS加密通信与流通信](https://razeen.me/post/how-to-use-grpc-in-golang-03.html)

在前面的两篇博客中，我们已经知道了如何利用gRPC建立简单RPC通信。但这样简单的实现有时候满足不了我们的业务需求。在一些场景中我们需要防止数据被劫持，或是一些场景中我们希望客户端与服务器不是简单的一问一答，而是建立起一个流式的RPC通信，那么该怎么做到呢？

<!--more-->

### TLS加密通信

TLS加密无非就是认证客户端与服务器，如果对SSL/TLS加密通信有所了解的童鞋都知道我们首先需要两张证书。

所以作为准备工作，我们首先要申请两张测试证书。一张客户端证书，一张服务器证书。

#### 生成测试证书

利用[MySSL测试证书生成工具](https://myssl.com/create_test_cert.html)我们可以很简单的生成两张证书，如下所示：

如图，填入域名生成一张服务器证书，然后将私钥，证书链，根证书都下载下来，保存到文件。

![](https://st.razeen.me/essay/image/grpc-demo3-001.png)

同样，生成一张客户端证书并保存。

![](https://st.razeen.me/essay/image/grpc-demo3-002.png)

#### 客户端与服务器TLS认证

在gRPC通信中，我们完成服务器认证与客户端认证主要使用的是grpc下的[credentials](https://godoc.org/google.golang.org/grpc/credentials)库。下面通过实例来看看怎么使用。

[代码实例]()

**服务端实现**

```go
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

	// 将根证书加入证书池
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
```

**客户端实现**

```go
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
		panic("fail to append test ca")
	}
    
    // 新建凭证
    // ServerName 需要与服务器证书内的通用名称一致
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
		log.Printf("%v", err)
	}

	log.Printf("Resp1:%+v", resp1)

	resp2, err := client.SayHelloWorld(context.Background(), &pb.HelloWorldRequest{
		Greeting: "Hello Server 2 !!",
	})
	if err != nil {
		log.Printf("%v", err)
	}

	log.Printf("Resp2:%+v", resp2)
}
```

从代码中，我们不难看出，主要是创建一个通信凭证(TransportCredentials)。利用`credentials`库的`NewTLS`方法从`tls`加载一个通信凭证用于通信。而在其中需要注意的是：

- 如果你使用的是自签发的证书，注意将根加入证书池。如果你使用的是可信CA签发的证书大部分不用添加，因为系统的可信CA库已经有了。如果没有成功添加, 在通信时会出现以下错误：

  >  rpc error: code = Unavailable desc = all SubConns are in TransientFailure, latest connection error: connection error: desc = "transport: authentication handshake failed: x509: certificate signed by unknown authority"

  或

  > rpc error: code = Unavailable desc = all SubConns are in TransientFailure, latest connection error: connection error: desc = "transport: authentication handshake failed: remote error: tls: bad certificate"

- 客户端凭证内 `ServerName` 需要与服务器证书内的通用名称一致，如果不一致会出现如下错误：

  > rpc error: code = Unavailable desc = all SubConns are in TransientFailure, latest connection error: connection error: desc = "transport: authentication handshake failed: x509: certificate is valid for server.razeen.me, not xxxxx"

之后，我们就可安心的通信了，在私钥不泄漏的情况下，基本不再担心数据劫持问题了。



**这里我想多说一句：**我们经常在提交代码时会直接 `git add .` ，这是个不好的习惯，有时后我们会将一些不必要的文件提交上去，特别是一些**证书**、**私钥**、**密码**之类的文件。



### 流式的RPC通信

流式PRC通信可以分为:

- 服务器端流式 RPC;

  客户端发送请求到服务器，拿到一个流去读取返回的消息序列。 客户端读取返回的流，直到里面没有任何消息。如：

  ```protobuf
  rpc ListHello(HelloWorldRequest) returns (stream HelloWorldResponse) {}
  ```

  

- 客户端流式 RPC;

  客户端写入一个消息序列并将其发送到服务器，同样也是使用流。一旦客户端完成写入消息，它等待服务器完成读取返回它的响应。如：

  ```protobuf
  rpc SayMoreHello(stream HelloWorldRequest) returns (HelloWorldResponse) {}
  ```

  

- 双向流式 RPC;

  双方使用读写流去发送一个消息序列。两个流独立操作，因此客户端和服务器可以以任意喜欢的顺序读写。如：

  ```protobuf
  rpc SayHelloChat(stream HelloWorldRequest) returns (stream HelloWorldRequest) {}
  ```



从上面的定义不难看出，用`stream`可以定义一个流式消息。下面我们就通过实例来演示一下流式通信的使用方法。

首先，我们将上面三个rpc server加入`.proto` , 并且生成新的`.pb.go`代码。

在生成的代码`hello_world.pb.go`中，我们可以看到客户端接口如下：

```go
type HelloWorldServiceClient interface {
	SayHelloWorld(ctx context.Context, in *HelloWorldRequest, opts ...grpc.CallOption) (*HelloWorldResponse, error)
	ListHello(ctx context.Context, in *HelloWorldRequest, opts ...grpc.CallOption) (HelloWorldService_ListHelloClient, error)
	SayMoreHello(ctx context.Context, opts ...grpc.CallOption) (HelloWorldService_SayMoreHelloClient, error)
	SayHelloChat(ctx context.Context, opts ...grpc.CallOption) (HelloWorldService_SayHelloChatClient, error)
}
```

服务端接口如下:

```go
// HelloWorldServiceServer is the server API for HelloWorldService service.
type HelloWorldServiceServer interface {
	SayHelloWorld(context.Context, *HelloWorldRequest) (*HelloWorldResponse, error)
	ListHello(*HelloWorldRequest, HelloWorldService_ListHelloServer) error
	SayMoreHello(HelloWorldService_SayMoreHelloServer) error
	SayHelloChat(HelloWorldService_SayHelloChatServer) error
}
```

在客户段的接口中，生成了`HelloWorldService_XXXXClient`接口类型。  在服务端的接口中，生成了`HelloWorldService_XXXXServer`接口类型。 我们再查看这些接口的定义，发现这这几个接口都是实现了以下几个方法中的数个：

```go
Send(*HelloWorldRequest) error
Recv() (*HelloWorldRequest, error)
CloseAndRecv() (*HelloWorldResponse, error)
grpc.ClientStream
```

看其名字，我们不难知道，流式RPC的使用，或者说流的收发也就离不开这几个方法了。下面我们通过几个实例来验证一下。



在服务端，我们实现这三个接口。

```go
// 服务器端流式 RPC, 接收一次客户端请求，返回一个流
func (s *SayHelloServer) ListHello(in *pb.HelloWorldRequest, stream pb.HelloWorldService_ListHelloServer) error {
	log.Printf("Client Say: %v", in.Greeting)

    // 我们返回多条数据
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

// 双向流式 RPC
func (s *SayHelloServer) SayHelloChat(stream pb.HelloWorldService_SayHelloChatServer) error {
	// 开一个协程去处理客户端数据
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

    // 向客户端写入多条数据
	stream.Send(&pb.HelloWorldRequest{Greeting: "SayHelloChat Server Say Hello 1"})
	time.Sleep(1 * time.Second)
	stream.Send(&pb.HelloWorldRequest{Greeting: "SayHelloChat Server Say Hello 2"})
	time.Sleep(1 * time.Second)
	stream.Send(&pb.HelloWorldRequest{Greeting: "SayHelloChat Server Say Hello 3"})
	time.Sleep(1 * time.Second)
	return nil
}
```



之后我们就可以在客户端分别请求这几个rpc服务。

```go
    // 服务器端流式 RPC;
    // 我们向服务器SayHello 
	recvListHello, err := client.ListHello(context.Background(), &pb.HelloWorldRequest{Greeting: "Hello Server List Hello"})
	if err != nil {
		log.Fatalf("ListHello err: %v", err)
	}

    // 服务器以流式返回
    // 直到 err==io.EOF时，表示接收完毕。
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
// Client Out:
// 2018/08/06 01:27:55 ListHello Server Resp: ListHello Reply Hello Server List Hello 1
// 2018/08/06 01:27:56 ListHello Server Resp: ListHello Reply Hello Server List Hello 2
// 2018/08/06 01:27:57 ListHello Server Resp: ListHello Reply Hello Server List Hello 3
// Server Out:
// 2018/08/06 01:27:55 Client Say: Hello Server List Hello


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

// Client Out:
// 2018/08/06 01:31:11 SayMoreHello Server Resp: SayMoreHello Recv Muti Greeting
// Server Out:
// 2018/08/06 01:31:11 SayMoreHello Client Say: SayMoreHello Hello Server 0
// 2018/08/06 01:31:11 SayMoreHello Client Say: SayMoreHello Hello Server 1
// 2018/08/06 01:31:11 SayMoreHello Client Say: SayMoreHello Hello Server 2


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
// Client Out:
// 2018/08/06 01:31:11 SayHelloChat Server Say: SayHelloChat Server Say Hello 1
// 2018/08/06 01:31:12 SayHelloChat Server Say: SayHelloChat Server Say Hello 2
// 2018/08/06 01:31:13 SayHelloChat Server Say: SayHelloChat Server Say Hello 3
// Server Out:
// 2018/08/06 01:31:11 SayHelloChat Client Say: SayHelloChat Hello Server 0
// 2018/08/06 01:31:11 SayHelloChat Client Say: SayHelloChat Hello Server 1
// 2018/08/06 01:31:11 SayHelloChat Client Say: SayHelloChat Hello Server 2
```

看了实例，是不是觉得很简单～。三种方式大同小异，只要掌握了怎么去收发流，怎么判断流的结束，基本就可以了。





好了，gRPC在Go中的使用三篇文章到这里也就结束了，如果博客中有错误或者你还有想知道的，记得留言哦。







