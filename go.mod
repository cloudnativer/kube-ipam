module kube-ipam

go 1.15

require (
	github.com/containernetworking/cni v0.8.1
	github.com/containernetworking/plugins v0.8.1
	github.com/coreos/bbolt v1.3.6 // indirect
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/coreos/go-iptables v0.6.0 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/safchain/ethtool v0.0.0-20210803160452-9aa261dae9b1 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/vishvananda/netlink v1.1.0 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.uber.org/zap v1.19.1 // indirect
	google.golang.org/grpc v1.41.0 // indirect
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.20.6
)

replace (
	github.com/coreos/bbolt v1.3.6 => go.etcd.io/bbolt v1.3.5
	google.golang.org/grpc v1.41.0 => google.golang.org/grpc v1.29.1
)
