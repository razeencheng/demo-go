# [Go学习笔记（十）老项目迁移 go module 大型灾难记录](https://razeencheng.com/post/accidents-of-migrating-to-go-modules.html)



最近在改造一个比较早期的一个项目，其中就涉及到用将原来 `Vendor` 管理依赖换成 `Go Modules` 来管理。 然而过程真是一波三折，在这里总结一下此次 `Go Modules` 改造中遇到的问题，以及解决方法。



<!--more-->



### 背景

- go version：

  ```bash
  $ go version
  go version go1.16.5 darwin/amd64
  ```

- 简化的 demo 如下,  很 “简单” 我们只要把 `hello world` 输出即可。

  ``` go
  package main
  
  import (
  	"github.com/coreos/etcd/pkg/transport"
  	"github.com/google/certificate-transparency-go/tls"
  	"github.com/qiniu/api.v7/auth/qbox"
  	"go.etcd.io/etcd/clientv3"
  	"google.golang.org/grpc"
  	"qiniupkg.com/x/log.v7"
  )
  
  func main() {
  
  	_ = transport.TLSInfo{}
  
  	_ = clientv3.WatchResponse{}
  
  	_, _ = clientv3.New(clientv3.Config{})
  
  	_ = qbox.NewMac("", "")
  
  	_ = tls.DigitallySigned{}
  
  	_ = grpc.ClientConn{}
  
  	log.Info("hello world")
  }
  ```
  
  

### 实战

直接初始化，并 tidy 一下。

```bash
$ go mod init demo-go/gomod
go: creating new go.mod: module demo-go/gomod
go: to add module requirements and sums:
        go mod tidy
   
$ go mod tidy
go: finding module for ...
demo-go/gomod imports
        qiniupkg.com/x/log.v7: module qiniupkg.com/x@latest found (v1.11.5), but does not contain package qiniupkg.com/x/log.v7
demo-go/gomod imports
        github.com/qiniu/api.v7/auth/qbox imports
        github.com/qiniu/x/bytes.v7/seekable: module github.com/qiniu/x@latest found (v1.11.5), but does not contain package github.com/qiniu/x/bytes.v7/seekable
demo-go/gomod imports
        go.etcd.io/etcd/clientv3 imports
        github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context: package github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context provided by github.com/coreos/etcd at latest version v2.3.8+incompatible but not at required version v3.3.10+incompatible
demo-go/gomod imports
        go.etcd.io/etcd/clientv3 imports
        github.com/coreos/etcd/Godeps/_workspace/src/google.golang.org/grpc: package github.com/coreos/etcd/Godeps/_workspace/src/google.golang.org/grpc provided by github.com/coreos/etcd at latest version v2.3.8+incompatible but not at required version v3.3.10+incompatible
demo-go/gomod imports
        go.etcd.io/etcd/clientv3 imports
        github.com/coreos/etcd/Godeps/_workspace/src/google.golang.org/grpc/credentials: package github.com/coreos/etcd/Godeps/_workspace/src/google.golang.org/grpc/credentials provided by github.com/coreos/etcd at latest version v2.3.8+incompatible but not at required version v3.3.10+incompatible
demo-go/gomod imports
        go.etcd.io/etcd/clientv3 imports
        github.com/coreos/etcd/storage/storagepb: package github.com/coreos/etcd/storage/storagepb provided by github.com/coreos/etcd at latest version v2.3.8+incompatible but not at required version v3.3.10+incompatible
```



好家伙，报错了。我们先看到前两行

1. `qiniupkg.com/x@latest`  中没有 `qiniupkg.com/x/log.v7`；
2. `github.com/qiniu/x@latest` 中没有 `github.com/qiniu/x/bytes.v7/seekable`；

这看起来应该是一个问题， `qiniupkg.com/x` 和`github.com/qiniu/x`  应该是同一个包，不同镜像。于是我到 Github 看一下 `@lastet` 版本的代码，确实没有`bytes.v7` 包了。人肉查找，最后在 `v1.7.8` 版本，我们找到了 `bytes.v7` 包。  



于是，我们可以指定一下版本。

```bash
go mod edit -replace qiniupkg.com/x=qiniupkg.com/x@v1.7.8
go mod edit -replace github.com/qiniu/x=github.com/qiniu/x@v1.7.8
```



继续往下看，接下来的几个问题是一类的，都是`etcd`导致的。 

意思是 `go.etcd.io/etcd/clientv3` 导入了 `github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context`, 同时 `github.com/coreos/etcd@v2.3.8`  中 提供了 `github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context` 。 但是，我们这里需要 `github.com/coreos/etcd@v3.3.10`, 而该版本并不提供  `github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context` 。

我们直接更新 etcd 到的 `v3.3.10` 试试。

```bash
go mod edit -replace go.etcd.io/etcd=go.etcd.io/etcd@v3.3.20+incompatible
```





我们再 ` go mod tidy` 下。

```bash
$ go mod tidy
go: demo-go/gomod imports
        go.etcd.io/etcd/clientv3 tested by
        go.etcd.io/etcd/clientv3.test imports
        github.com/coreos/etcd/auth imports
        github.com/coreos/etcd/mvcc/backend imports
        github.com/coreos/bbolt: github.com/coreos/bbolt@v1.3.6: parsing go.mod:
        module declares its path as: go.etcd.io/bbolt
                but was required as: github.com/coreos/bbolt
```

这个错误和鸟窝这篇 [Etcd使用go module的灾难](https://colobu.com/2020/04/09/accidents-of-etcd-and-go-module/)一致，`go.etcd.io/bbolt` 和 `github.com/coreos/bbolt` 包名不一致，我们替换一下。

``` bash
go mod edit -replace github.com/coreos/bbolt@v1.3.6=go.etcd.io/bbolt@v1.3.6
```





继续，`go mod tidy` 一下。

```bash
$ go mod tidy
...
demo-go/gomod imports
        go.etcd.io/etcd/clientv3 imports
        github.com/coreos/etcd/clientv3/balancer: module github.com/coreos/etcd@latest found (v2.3.8+incompatible), but does not contain package github.com/coreos/etcd/clientv3/balancer
demo-go/gomod imports
        go.etcd.io/etcd/clientv3 imports
        github.com/coreos/etcd/clientv3/balancer/picker: module github.com/coreos/etcd@latest found (v2.3.8+incompatible), but does not contain package github.com/coreos/etcd/clientv3/balancer/picker
demo-go/gomod imports
        go.etcd.io/etcd/clientv3 imports
        github.com/coreos/etcd/clientv3/balancer/resolver/endpoint: module github.com/coreos/etcd@latest found (v2.3.8+incompatible), but does not contain package github.com/coreos/etcd/clientv3/balancer/resolver/endpoint
demo-go/gomod imports
        go.etcd.io/etcd/clientv3 imports
        github.com/coreos/etcd/clientv3/credentials: module github.com/coreos/etcd@latest found (v2.3.8+incompatible), but does not contain package github.com/coreos/etcd/clientv3/credentials
demo-go/gomod imports
        go.etcd.io/etcd/clientv3 tested by
        go.etcd.io/etcd/clientv3.test imports
        github.com/coreos/etcd/integration imports
        github.com/coreos/etcd/proxy/grpcproxy imports
        google.golang.org/grpc/naming: module google.golang.org/grpc@latest found (v1.39.0), but does not contain package google.golang.org/grpc/naming
```

好家伙，又是`etcd`。 仔细一看，我们导入了`github.com/coreos/etcd` 和 `go.etcd.io/etcd` 两个版本`etcd`, 我们前面只替换了一个。现在我们把另外一个也替换了。

```bash
go mod edit -replace github.com/coreos/etcd=github.com/coreos/etcd@v3.3.20+incompatible
```





再`go mod tidy`下，这个错误没有了，但还有个`grpc`的错误，继续找原因。原来是` google.golang.org/grpc` `v1.39.0` 版本没有` google.golang.org/grpc/naming` 包。 上 Github 仓库， 找了一下历史版本，`v1.29.1`上是有这个包的，我们继续替换。

```bash
go mod edit -replace google.golang.org/grpc=google.golang.org/grpc@v1.29.1
```





这下，终于，`go mod tidy`通过了，可以开心的输出`hello world` 了。





然而，

``` bash
$ go run main.go
# github.com/coreos/etcd/clientv3/balancer/resolver/endpoint
../../../go/pkg/mod/github.com/coreos/etcd@v3.3.20+incompatible/clientv3/balancer/resolver/endpoint/endpoint.go:114:78: undefined: resolver.BuildOption
../../../go/pkg/mod/github.com/coreos/etcd@v3.3.20+incompatible/clientv3/balancer/resolver/endpoint/endpoint.go:182:31: undefined: resolver.ResolveNowOption
# github.com/coreos/etcd/clientv3/balancer/picker
../../../go/pkg/mod/github.com/coreos/etcd@v3.3.20+incompatible/clientv3/balancer/picker/err.go:37:44: undefined: balancer.PickOptions
../../../go/pkg/mod/github.com/coreos/etcd@v3.3.20+incompatible/clientv3/balancer/picker/roundrobin_balanced.go:55:54: undefined: balancer.PickOptions
```

意不意外，惊不惊喜！!

原来`etcd`包依赖了`grpc`的`resolver`包，但我导入的`v1.29.1`版本的`grpc`是没有这个包的。到 `grpc`[仓库](https://github.com/grpc/grpc-go/blob/v1.26.0/resolver/resolver.go) 挨个版本看了一下，确实只有`v1.26.0`版本才声明了`type BuildOption` 。于是，我们再次使用替换大法。

```bash
go mod edit -replace google.golang.org/grpc=google.golang.org/grpc@v1.26.0
```



再次`tidy`, 运行！ 终于，看到了久违的`hello world!`

```bash
$ go run main.go
2021/07/20 12:27:09.642431 [INFO] /Users/razeen/wspace/github/demo-go/gomod/main.go:26: hello world
```





### 总结

#### 项目规范

现在我们回过头看下这个 demo 项目，其实很有问题。

``` go
	"github.com/coreos/etcd/pkg/transport"
	"github.com/google/certificate-transparency-go/tls"
	"github.com/qiniu/api.v7/auth/qbox"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"qiniupkg.com/x/log.v7"
```

`etcd` 和 `qiniupkg`的包完全可以统一，只导入一种！而且，后来我们发现`log.v7`这包也是意外导入的....

这也是在改造我们一些老的项目时遇到的问题，以前用`vendor` `go get` 没有注意到这些问题，这是需要提前规范的。



#### 看懂 `go.mod`

我们来简单看一下，经历各种坎坷后，得出的`go.mod` 文件。

```  go
module demo-go/gomod

go 1.16

replace qiniupkg.com/x => qiniupkg.com/x v1.7.8

replace github.com/qiniu/x => github.com/qiniu/x v1.7.8

replace go.etcd.io/etcd => go.etcd.io/etcd v3.3.20+incompatible

replace github.com/coreos/bbolt v1.3.6 => go.etcd.io/bbolt v1.3.6

replace github.com/coreos/etcd => github.com/coreos/etcd v3.3.20+incompatible

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/coreos/bbolt v1.3.6 // indirect
	github.com/coreos/etcd v3.3.10+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/google/certificate-transparency-go v1.1.1
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/qiniu/api.v7 v7.2.5+incompatible
	github.com/qiniu/x v0.0.0-00010101000000-000000000000 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/etcd v0.0.0-20200513171258-e048e166ab9c
	google.golang.org/grpc v1.29.1
	qiniupkg.com/x v0.0.0-00010101000000-000000000000
	sigs.k8s.io/yaml v1.2.0 // indirect
)
```



我们先看一个常见的这几个[指令](https://golang.org/ref/mod#go-mod-file-module)，

- `module` 定义主模块的路径；
- `go` 编写该`mod`文件时的go版本；
- `require` 声明给定模块依赖项的最低要求版本;
- `replace` 手动指定的依赖模块 (可以替换全部的版本、指定的版本、本地的版本等等 )；



还有就是 `v3.3.20+incompatible` 后面的 `+incompatible` , 这是指兼容的版本，指依赖库的版本是`v2` 或以上，但`go.mod`和 依赖库路径 没有按照官方指定的方式命名，会加上这个。



`v0.0.0-00010101000000-000000000000` 这是一个伪版本，在和 不兼容 module 或 标记的版本不可用的时候，回打上这个伪版本。



`// indirect` 这指明这些不是我们直接引用的依赖。



除此之外，以下指令也可了解一下。

``` bash
# 查看当前模块以及所有的依赖模块
go list -m all

# 查看某个模块的以及打标签的版本
go list -m -versions go.etcd.io/etcd

# 升级特定的包
go get xx@version 升级特定的包

# 了解为什么需要模块
go mod why -m all  

# 为什么需要指定（google.golang.org/grpc）的模块
go mod why -m google.golang.org/grpc
```



更多可以细读[官方文档](https://golang.org/ref/mod#incompatible-versions)，感谢阅读。





### 参考

- [Using Go Modules](https://blog.golang.org/using-go-modules)
- [Minimal Version Selection](https://research.swtch.com/vgo-mvs)
- [跳出Go module的泥潭](https://colobu.com/2018/08/27/learn-go-module/)
- [Etcd使用go module的灾难](https://colobu.com/2020/04/09/accidents-of-etcd-and-go-module/)
- [浅谈Go Modules原理](https://duyanghao.github.io/golang-module/)

