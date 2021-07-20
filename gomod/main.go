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
