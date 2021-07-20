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
