package kipam

import (
	"fmt"
)

func OutputCniConfig(outputstr string) {
	fmt.Println(`/etc/cni/net.d/1-kube-ipam.conf 
----------------------------------------------------
{
        "cniVersion":"0.3.1",
        "name": "k8snetwork", `)
        fmt.Println("        \"type\": \""+outputstr+"\",\"")
        fmt.Println(`        "master": "eth0",
        "ipam": {
                "name": "kube-subnet",
                "type": "kube-ipam",
		"kubeConfig": "/etc/kubernetes/pki/kubectl.kubeconfig",
                "etcdConfig": {
                        "etcdURL": "https://192.168.122.11:2379",
                        "etcdCertFile": "/etc/kubernetes/ssl/kubernetes.pem",
                        "etcdKeyFile": "/etc/kubernetes/ssl/kubernetes-key.pem",
                        "etcdTrustedCAFileFile": "/etc/kubernetes/ssl/k8s-root-ca.pem"
                },
                "subnet": "10.244.0.0/16",
                "rangeStart": "10.244.0.10",
                "rangeEnd": "10.244.0.200",
                "gateway": "10.244.0.1",
                "routes": [{
                        "dst": "0.0.0.0/0"
                }],
                "resolvConf": "/etc/resolv.conf"
        }
}
----------------------------------------------------
    `)
}
