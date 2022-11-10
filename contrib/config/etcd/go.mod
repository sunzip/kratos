module github.com/sunzip/kratos/contrib/config/etcd/v2

go 1.16

require (
	github.com/sunzip/kratos/v2 v2.2.0
	go.etcd.io/etcd/client/v3 v3.5.0
	google.golang.org/grpc v1.44.0
)

replace github.com/sunzip/kratos/v2 => ../../../
