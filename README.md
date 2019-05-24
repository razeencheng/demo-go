# 学习golang的一些Demo

*更多内容请看博客  👉 https://razeencheng.com*

### Demo 一览

1. [Go学习笔记(二) | 我对recover的一点误解](https://razeen.me/post/daily-go-recover.html)  👉 [recover](https://github.com/razeencheng/demo-go/tree/master/recover)
2. [Go学习笔记(三) | 怎么写Go基准测试（性能测试）](https://razeen.me/post/go-how-to-write-benchmark.html) 👉 [benchmark](https://github.com/razeencheng/demo-go/tree/master/benchmark)
3. [gRPC在Go中的使用（一）Protocol Buffers语法与相关使用](https://razeen.me/post/how-to-use-grpc-in-golang-01.html)  👉 [grpc/demo1](https://github.com/razeencheng/demo-go/tree/master/grpc/demo1)
4. [gRPC在Go中的使用（二）gRPC实现简单通讯](https://razeen.me/post/how-to-use-grpc-in-golang-02.html)  👉 [grpc/demo2](https://github.com/razeencheng/demo-go/tree/master/grpc/demo2)
5. [gRPC在Go中的使用（三）gRPC实现TLS加密通信与流模式](https://razeen.me/post/how-to-use-grpc-in-golang-03.html)  👉 [grpc/demo3](https://github.com/razeencheng/demo-go/tree/master/grpc/demo3)

6. json tag 使用 👉 [json/tag](https://github.com/razeencheng/demo-go/tree/master/json/tag)

7. [gin文件上传与下载](https://newb.razeen.me/gin-file-download-and-upload/) 👉 [gin/gin.go](https://github.com/razeencheng/demo-go/blob/master/gin/gin.go)

8. [Go学习笔记(六) 使用swaggo自动生成Restful API文档](https://razeen.me/post/go-swagger.html) 👉 [swaggo-gin](https://github.com/razeencheng/demo-go/tree/master/swaggo-gin)

9. [Go学习笔记(七) | 理解并实现 OAuth 2.0](https://razeen.me/post/oauth2-protocol-details.html) 👉  [oauth2](https://github.com/razeencheng/demo-go/tree/master/oauth2)

10. [如何用Go调用Windows API](https://razeencheng.com/post/breaking-all-the-rules-using-go-to-call-windows-api.html)  👉 [windows_api](https://github.com/razeencheng/demo-go/tree/master/windows_api)


### 目录

````
.
├── README.md
├── benchmark    
│   ├── main.go
│   └── main_test.go
├── gin
│   ├── gin.go
├── grpc
│   ├── demo1
│   │   └── helloworld
│   │       ├── hello_world.pb.go
│   │       └── hello_world.proto
│   ├── demo2
│   │   ├── client
│   │   │   └── client.go
│   │   ├── helloworld
│   │   │   ├── hello_world.pb.go
│   │   │   └── hello_world.proto
│   │   └── server
│   │       └── server.go
│   └── demo3
│       ├── client
│       │   ├── certs
│       │   └── client.go
│       ├── helloworld
│       │   ├── hello_world.pb.go
│       │   └── hello_world.proto
│       └── server
│           ├── certs
│           └── server.go
├── json
│   └── tag
│       └── tag.go
│── recover
│   └── recover.go
├── oauth2
│   ├── README.md
│   ├── main.go
│   └── public
│       ├── index.html
│       └── welcome.html
├── recover
│   ├── README.md
│   └── recover.go
├── swaggo-gin
│   ├── Makefile
│   ├── README.md
│   ├── doc.go
│   ├── docs
│   │   ├── docs.go
│   │   └── swagger
│   │       ├── swagger.json
│   │       └── swagger.yaml
│   ├── handle.go
│   └── main.go
│   
└── windows_api
    ├── README.md
    └── main_windows.go

````
