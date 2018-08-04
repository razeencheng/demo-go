# [gRPC在Go中的使用（一）Protocol Buffers语法与相关使用](https://razeen.me/post/how-to-use-grpc-in-golang-01.html)

Desc:protobuf语法介绍，怎么写proto文件，grpc的使用入门

在gRPC官网用了一句话来介绍:“一个高性能、开源的通用RPC框架”，同时介绍了其四大特点：

* 定义简单
* 支持多种编程语言多种平台
* 快速启动和缩放
* 双向流媒体和集成身份验证 

<!--more-->

在`gRPC在go中使用`系列中，关于其简介与性能我就不多介绍，相信在社区也有很多关于这些的讨论。这里我主要从三个层次来总结我以往在Go中使用gRPC的一些经验，主要分为：

1. Protocol Buffers语法与相关使用
2. gRPC实现简单通讯
3. gRPC服务认证与双向流通讯

**注:下面Protocol Buffers简写protobuf.*


这篇我们先介绍protobuf的相关语法、怎么书写`.proto`文件以及go代码生成。

### 简介

要熟练的使用GRPC，protobuf的熟练使用必不可少。

gRPC使用[protobuf](https://github.com/google/protobuf)来定义服务。protobuf是由Google开发的一种数据序列化协议，可以把它想象成是XML或JSON格式，但更小，更快更简洁。而且一次定义，可生成多种语言的代码。

### 定义

首先我们需要编写一些`.proto`文件，定义我们在程序中需要处理的结构化数据。我们直接从一个实例开始讲起，下面是一个proto文件：


``` protobuf
syntax = "proto3";

option go_package = "github.com/razeencheng/demo-go/grpc/demo1/helloworld";

package helloworld;

import "github.com/golang/protobuf/ptypes/any/any.proto";

message HelloWorldRequest {
  string greeting = 1;
  map<string, string> infos  = 2;
}

message HelloWorldResponse {
  string reply = 1;
  repeated google.protobuf.Any details = 2;
}

service HelloWorldService {
  rpc SayHelloWorld(HelloWorldRequest) returns (HelloWorldResponse){}
}
```


#### 版本 

文件的开头`syntax="proto3"`也就指明版本，主要有`proto2`与`proto3`,他们在语法上有一定的差异，我这里主要使用的是后者。

#### 包名

第二行，指定生成go文件的包名，可选项，默认使用第三行包名。

第三行，包名。

#### 导包

第四行，类似你写go一样，protobuf也可以导入其他的包。

#### 消息定义

后面message开头的两个结构就是我们需要传递的消息类型。所有的消息类型都是以`message`开始，然后定义类型名称。结构内字段的定义为`字段规则 字段类型 字段名=字段编号`

- 字段规则主要有 `singular`和`repeated`。如其中`greeting`和`reply`的字段规则为`singular`,允许该消息中出现0个或1个该字段(但不能超过一个)，而像`details`字段允许重复任意次数。其实对应到go里面也就是基本类型和切片类型。
- 字段类型，下表是proto内类型与go类型的对应表。

| .proto Type | Notes                                                        | Go Type |
| ----------- | ------------------------------------------------------------ | ------- |
| double      |                                                              | float64 |
| float       |                                                              | float32 |
| int32       | 使用可变长度编码。 无效编码负数 - 如果您的字段可能具有负值， 请改用sint32。 | int32   |
| int64       | 使用可变长度编码。 无效编码负数 - 如果您的字段可能具有负值，请改用sint64。 | int64   |
| uint32      | 使用可变长度编码。                                           | uint32  |
| uint64      | 使用可变长度编码。                                           | uint64  |
| sint32      | 使用可变长度编码。 带符号的int值。 这些比常规的int32更有效地编码负数。 | int32   |
| sint64      | 使用可变长度编码。 带符号的int值。 这些比常规的int64更有效地编码负数。 | int64   |
| fixed32     | 总是四个字节。 如果值通常大于228，则比uint32效率更高。       | uint32  |
| fixed64     | 总是八个字节。 如果值通常大于256，则会比uint64更高效。       | uint64  |
| sfixed32    | 总是四个字节。                                               | int32   |
| sfixed64    | 总是八个字节。                                               | int64   |
| bool        |                                                              | bool    |
| string      | 字符串必须始终包含UTF-8编码或7位ASCII文本。                  | string  |
| bytes       | 可能包含任何字节序列。                                       | []byte  |

看到这里你也许会疑惑，go里面的切片，map，接口等类型我怎么定义呢？别急，下面一一替你解答。

1.map类型，`HelloWorldRequest`的`infos`就是一个map类型，它的结构为`map<key_type, value_type> map_field = N`  但是在使用的时候你需要注意map类型不能`repetead`。

2.切片类型，我们直接定义其规则为`repeated`就可以了。就像`HelloWorldResponse`中的`details`字段一样，它就是一个切片类型。那么你会问了它是什么类型的切片？这就看下面了~

3.接口类型在proto中没有直接实现，但在[google/protobuf/any.proto](https://github.com/golang/protobuf/blob/master/ptypes/any/any.proto)中定义了一个`google.protobuf.Any`类型，然后结合[protobuf/go](https://github.com/golang/protobuf/blob/master/ptypes/any.go)也算是曲线救国了~

- 字段编号

  最后的1，2代表的是每个字段在该消息中的唯一标签，在与消息二进制格式中标识这些字段，而且当你的消息在使用的时候该值不能改变。1到15都是用一个字节编码的，通常用于标签那些频繁发生修改的字段。16到2047用两个字节编码，最大的是2^29-1(536,870,911)，其中19000-19999为预留的，你也不可使用。



#### 服务定义

如果你要使用RPC(远程过程调用)系统的消息类型，那就需要定义RPC服务接口，protobuf编译器将会根据所选择的不同语言生成服务接口代码及存根。就如：

```
service HelloWorldService {
  rpc SayHelloWorld(HelloWorldRequest) returns (HelloWorldResponse){}
}
```

protobuf编译器将产生一个抽象接口`HelloWorldService`以及一个相应的存根实现。存根将所有的调用指向RpcChannel(SayHelloWorld)，它是一个抽象接口，必须在RPC系统中对该接口进行实现。具体如何使用，将在下一篇博客中详细介绍。


### 生成Go代码

#### 安装protoc

首先要安装`protoc`,可直接到[这里](https://github.com/google/protobuf/releases/tag/v3.0.0)下载二进制安装到 `$PATH`里面，也可以直接下载源码编译。除此之外，你还需要安装go的proto插件`protoc-gen-go`。

```Go
// mac terminal
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
// win powershell
go get -u github.com/golang/protobuf/proto
go get -u github.com/golang/protobuf/protoc-gen-go
```

#### 生成go代码

接下来，使用`protoc`命令即可生成。

```Bash
### mac terminal
protoc -I ${GOPATH}/src --go_out=plugins=grpc:${GOPATH}/src ${GOPATH}/src/github.com/razeencheng/demo-go/grpc/demo1/helloworld/hello_world.proto
### win powershell
protoc -I $env:GOPATH\src --go_out=plugins=grpc:$env:GOPATH\src $env:GOPATH\src\github.com\razeencheng\demo-go\grpc\demo1\helloworld\hello_world.proto
```

如上所示 `-I`指定搜索proto文件的目录,`--go_out=plugins=grpc:`指定生成go代码的文件夹，后面就是需要生成的proto文件路径。

>  *注意：* 如果你使用到了其他包的结构，`-I`需要将该资源包括在内。
>
> 例如我导入了`github.com/golang/protobuf/ptypes/any/any.proto` 我首先需要
>
> `go get -u github.com/golang/protobuf`获取该包，然后在使用时资源路径(`-I`)直接为`GOPATH\src`。

最后生成的`hello-world.pb.go`文件。内容大概如下图所示，点[这里](https://github.com/razeencheng/demo-go/blob/master/grpc/demo1/helloworld/hello_world.pb.go)可查看全部。

<img src="https://st.razeen.me/essay/image/go/grpc-001.png" width="600px" height="200px">

<img src="https://st.razeen.me/essay/image/go/grpc-002.png" width="600px" height="200px">

图中我们可以看到两个`message`对应生成了两个结构体，同时会生成一些序列化的方法等。

<img src="https://st.razeen.me/essay/image/go/grpc-003.png" width="600px" height="200px">

<img src="https://st.razeen.me/essay/image/go/grpc-004.png" width="600px" height="200px">

而定义的`service`则是生成了对应的`client`与`server`接口，那么这到底有什么用？怎么去用呢？[下一篇博客](https://razeen.me/post/how-to-use-grpc-in-golang-02.html)将为你详细讲解~


看到这，我们简单的了解一下protobuf语法，如果你想了解更多，点[这里](https://developers.google.com/protocol-buffers/docs/proto3)看官方文档。
