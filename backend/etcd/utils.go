package etcd

import (
	"kube-ipam/backend/allocator"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
)

func connectStore(etcdConfig *allocator.EtcdConfig) (*clientv3.Client, error) {

	var etcdClient *clientv3.Client
	var err error
	if strings.HasPrefix(etcdConfig.EtcdURL, "https") {
		etcdClient, err = connectWithTLS(etcdConfig.EtcdURL, etcdConfig.EtcdCertFile, etcdConfig.EtcdKeyFile, etcdConfig.EtcdTrustedCAFileFile)
	} else {
		etcdClient, err = connectWithoutTLS(etcdConfig.EtcdURL)
	}

	return etcdClient, err
}

/*
	ETCD Related
*/
func connectWithoutTLS(url string) (*clientv3.Client, error) {
	etcdUrl := strings.Split(url, ",")
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdUrl,
		DialTimeout: 5 * time.Second,
	})
	return cli, err
}

func connectWithTLS(url, cert, key, trusted string) (*clientv3.Client, error) {
	etcdUrl := strings.Split(url, ",")
	tlsInfo := transport.TLSInfo{
		CertFile:      cert,
		KeyFile:       key,
		TrustedCAFile: trusted,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdUrl,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	return cli, err
}
