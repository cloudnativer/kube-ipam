

# Use `kube-ipam` to fix the IP address of pod in `calico` network environment



## [1] Installing and configuring calico CNI

### 1.1 Enable and configure directory

Please make sure that your kubelet takes the three correct parameters `network-plugin`, `cni-conf-dir` and `cni-bin-dir`. An example configuration of kubelet is given below:

```
# cat /etc/systemd/system/kubelet.service 
...
ExecStart=/usr/local/bin/kubelet \
...
  --network-plugin=cni \
  --cni-conf-dir=/etc/cni/net.d \
  --cni-bin-dir=/opt/cni/bin/ \
...
```

### 1.2 Install kube-ipam

`kube-ipam` binary program files can be obtained by <a href="docs/download.md">download</a> or <a href="docs/build.md">compile</a>, and copy the kube-ipam binary to the `/opt/cni/bin/` directory
```
# wget https://github.com/cloudnativer/kube-ipam/releases/download/v0.2.0/kube-ipam-v0.2.0-x86.tgz
# tar -zxvf kube-ipam-v0.2.0-x86.tgz
# mv kube-ipam-v0.2.0-x86/kube-ipam /opt/cni/bin/kube-ipam
```

### 1.3 Installing calico CNI

There is an example <a href="../yaml/calico.yaml">calico.yaml</a> in the yaml directory. We use this <a href="../yaml/calico.yaml">calico.yaml</a> to install calico CNI. Please note the settings of `kubeConfig`, `etcdConfig`, `subnet` and other parameters. `subnet` and `CALICO_IPV4POOL_CIDR` parameters should be set to the same value.

```
...

                "kubeConfig": "/etc/kubernetes/ssl/kube.kubeconfig",
                "etcdConfig": {
                        "etcdURL": "https://192.168.1.50:2379,https://192.168.1.58:2379,https://192.168.1.63:2379",
                        "etcdCertFile": "/etc/kubernetes/ssl/kubernetes.pem",
                        "etcdKeyFile": "/etc/kubernetes/ssl/kubernetes-key.pem",
                        "etcdTrustedCAFileFile": "/etc/kubernetes/ssl/k8s-root-ca.pem"
                },
                "subnet": "10.244.0.0/16",
                "rangeStart": "10.244.0.10",
                "rangeEnd": "10.244.0.200",
...

            - name: CALICO_IPV4POOL_CIDR
              value: "10.244.0.0/16"

...

```


Install calico CNI using the `kubectl apply -f ` command:


```
#
# kubectl apply -f yaml/calico.yaml
  ...

```

Use the `kubectl get` and `calicoctl node status` command to view the operation of calico CNI:


```
# 
# kubectl get pod -o wide -n kube-system | grep calico 
NAME                                 READY   STATUS       RESTARTS   AGE    IP             NODE
calico-kube-controllers-6457d89859   1/1     Running      0          155m   10.244.4.136    192.168.56.83
calico-node-pfjdc                    1/1     Running      0          17h    192.168.56.82   192.168.56.82
calico-node-t8nrb                    1/1     Running      0          17h    192.168.56.83   192.168.56.83
calico-node-tf69d                    1/1     Running      0          17h    192.168.56.81   192.168.56.81
#
# calicoctl node status
Calico process is running.

IPv4 BGP status
+---------------+-------------------+-------+----------+-------------+
| PEER ADDRESS  |     PEER TYPE     | STATE |  SINCE   |    INFO     |
+---------------+-------------------+-------+----------+-------------+
| 192.168.56.81 | node-to-node mesh | up    | 06:24:41 | Established |
| 192.168.56.82 | node-to-node mesh | up    | 06:25:09 | Established |
| 192.168.56.83 | node-to-node mesh | up    | 06:25:09 | Established |
+---------------+-------------------+-------+----------+-------------+

IPv6 BGP status
No IPv6 peers found.

#
#
```

### 1.4 About /etc/cni/net.d config

After calico CNI is installed, you will see the following configuration information in `/etc/cni/net.d/10-calico.conflist` file:

```

#
# cat /etc/cni/net.d/10-calico.conflist 
{
  "name": "k8s-pod-network",
  "cniVersion": "0.3.1",
  "plugins": [
    {
      "type": "calico",
      "log_level": "info",
      "log_file_path": "/var/log/calico/cni/cni.log",
      "datastore_type": "kubernetes",
      "nodename": "192.168.56.82",
      "mtu": 0,
      "ipam": {
                "name": "kube-subnet",
                "type": "kube-ipam",
                "kubeConfig": "/etc/kubernetes/ssl/kube.kubeconfig",
                "etcdConfig": {
                        "etcdURL": "https://192.168.1.50:2379,https://192.168.1.58:2379,https://192.168.1.63:2379",
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

#
#
```


## [2] Create fixed IP and random IP pod

### 2.1 Create fixed IP pod

Next, let's create a fixed IP Pod:

```
#
# cat fixed-ip-test.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: fixed-ip-test
  name: fixed-ip-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: fixed-ip-test
  strategy: {}
  template:
    metadata:
      labels:
        app: fixed-ip-test
      annotations:
        kube-ipam.ip: "10.244.0.216"
        kube-ipam.netmask: "255.255.0.0"
        kube-ipam.gateway: "10.244.0.1"
    spec:
      containers:
      - image: nginx:latest
        imagePullPolicy: IfNotPresent 
        name: nginx
        resources: {}
#
#
```

Create a fixed IP pod using the `kubectl apply -f` command:

```
#
#kubectl apply -f fixed-ip-test.yaml
 deployment.apps/fixed-ip-test configured

#
```

### 2.3 Create random IP pod

Next, let's create a random IP Pod:

```
#
# random-ip-test.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: random-ip-test
  name: random-ip-test
spec:
  replicas: 2
  selector:
    matchLabels:
      app: random-ip-test
  strategy: {}
  template:
    metadata:
      labels:
        app: random-ip-test
    spec:
      containers:
      - image: nginx:latest
        imagePullPolicy: IfNotPresent 
        name: nginx
        resources: {}

#
#
```

Create a random IP pod using the `kubectl apply -f` command:

```
#
#kubectl apply -f random-ip-test.yaml
 deployment.apps/random-ip-test configured

#
```


## [3] Verify the fixation of pod IP

### 3.1 View current pod status

Now you can see that there are two random-ip pod and one fixed-ip pod:

```
#
# kubectl get pod -o wide 
NAME                             READY   STATUS    RESTARTS   AGE    IP             NODE  
random-ip-test-59df9c4fdd-hnfxm  1/1     Running   0          128m   10.244.0.23   192.168.56.82
random-ip-test-7b75cbc65-ld9kb   1/1     Running   0          130m   10.244.0.24   192.168.56.83
fixed-ip-test-88554b798-xcfpb    1/1     Running   0          131m   10.244.0.216  192.168.56.81
#
#
```

### 3.2 After rescheduling, the pod IP remains unchanged

Use the `kubectl delete` command to delete fixed-ip pod, and kuberntes will automatically start a new fixed IP test pod:

```
# kubectl delete pod fixed-ip-test-88554b798-xcfpb 
#
# kubectl get pod -o wide                  
NAME                             READY   STATUS    RESTARTS   AGE    IP             NODE
random-ip-test-59df9c4fdd-hnfxm  1/1     Running   0          128m   10.244.0.23   192.168.56.82
random-ip-test-7b75cbc65-ld9kb   1/1     Running   0          130m   10.244.0.24   192.168.56.83
fixed-ip-test-88554b798-8ukle    1/1     Running   0          1m     10.244.0.216  192.168.56.82
#
```

At this time, the IP address of the newly started fixed-ip-test-88554b798-8ukle is still 10.244.0.216.



