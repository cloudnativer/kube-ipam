package kipam

import (
	"fmt"
)

func OutputCniConfig(outputstr string) {
    if outputstr == "calico" {
        fmt.Println(`/etc/cni/net.d/1-kube-ipam.conf
----------------------------------------------------
{
  "name": "k8s-pod-network",
  "cniVersion": "0.3.1",
  "plugins": [
    {
      "type": "calico",
      "log_level": "info",
      "log_file_path": "/var/log/calico/cni/cni.log",
      "datastore_type": "kubernetes",
      "nodename": "192.168.56.83",
      "mtu": 0,
      "ipam": {
                "name": "kube-subnet",
                "type": "kube-ipam",
                "kubeConfig": "/etc/kubernetes/pki/kubectl.kubeconfig",
                "etcdConfig": {
                        "etcdURL": "https://192.168.1.50:2379,https://192.168.1.58:2379,https://192.168.1.63:2379",
                        "etcdCertFile": "/etc/kubernetes/pki/kubernetes.pem",
                        "etcdKeyFile": "/etc/kubernetes/pki/kubernetes-key.pem",
                        "etcdTrustedCAFileFile": "/etc/kubernetes/pki/k8s-root-ca.pem"
                },
                "subnet": "10.244.0.0/16",
                "rangeStart": "10.244.0.10",
                "rangeEnd": "10.244.0.200",
                "gateway": "10.244.0.1",
                "routes": [{
                        "dst": "0.0.0.0/0"
                }],
                "resolvConf": "/etc/resolv.conf"
      },
      "policy": {
          "type": "k8s"
      },
      "kubernetes": {
          "kubeconfig": "/etc/cni/net.d/calico-kubeconfig"
      }
    },
    {
      "type": "portmap",
      "snat": true,
      "capabilities": {"portMappings": true}
    },
    {
      "type": "bandwidth",
      "capabilities": {"bandwidth": true}
    }
  ]
}
----------------------------------------------------
    `)
    } else {
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
                        "etcdURL": "https://192.168.1.50:2379,https://192.168.1.58:2379,https://192.168.1.63:2379",
                        "etcdCertFile": "/etc/kubernetes/pki/kubernetes.pem",
                        "etcdKeyFile": "/etc/kubernetes/pki/kubernetes-key.pem",
                        "etcdTrustedCAFileFile": "/etc/kubernetes/pki/k8s-root-ca.pem"
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
}
