# 基于kube-ipam与Multus实现Web和数据库分层网络安全访问架构



# 基本概况
## kube-ipam与Multus概述
Kube-ipam支持给kubernetes集群中的Pod固定IP地址。一些场景往往对IP地址有依赖，需要使用固定IP地址的Pod，可以使用kube-ipam轻松解决这类问题。例如，mysql主从架构的时候，主database与从database之间的同步；例如keepalived做集群HA的时候，两个节点之间检测通信等；例如某些安全防护设备，需要基于IP地址进行网络安全访问策略限制的场景等。
Multus-CNI支持同时添加多个网络接口到kubernetes环境中的Pod。这样的部署方式有利于安全人员把应用网络和数据库等多个网络区域进行相互隔离，有效控制容器集群网络架构。

## 网络分层架构设计
 
<br>

![kube-ipam](docs/images/Networksecuritylayering.jpg)

<br>

上图中显示了每个Pod具有2个接口：eth0、net1。eth0作为外界用户访问web pod的网络接口；而net1是附加的容器网卡，作为web Pod到database Pod的内部网络通信。



# 安装CNI插件

## 安装cni plugin和flannel
如果你使用的是kube-install安装的k8s，那么这两步可以省略：

安装cni plugin
```
# wget https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz
# tar -zxvf cni-plugins-linux-amd64-v0.9.1.tgz -C /opt/cni/bin/
```


安装flanneld
创建flanneld所需的subnet网段

```
# etcdctl --endpoints=https://192.168.122.11:2379,https://192.168.122.12:2379,https://192.168.122.13:2379  --ca-file=/etc/kubernetes/ssl/k8s-root-ca.pem --cert-file=/etc/kubernetes/ssl/kubernetes.pem --key-file=/etc/kubernetes/ssl/kubernetes-key.pem set /kubernetes/network/config '{"Network":"10.244.0.0/16", "SubnetLen":24, "Backend":{"Type":"vxlan"}}'
# wget https://github.com/flannel-io/flannel/releases/download/v0.11.0/flannel-v0.11.0-linux-amd64.tar.gz
# tar -zxvf flannel-v0.11.0-linux-amd64.tar.gz -C /opt/cni/bin/
```

编辑 /etc/systemd/system/flanneld.service

```
# cat /etc/systemd/system/flanneld.service
[Unit]
Description=Flanneld overlay address etcd agent
After=network.target
After=network-online.target
Wants=network-online.target
After=etcd.service
Before=docker.service

[Service]
Type=notify
ExecStart=/opt/cni/bin/flanneld \
 -etcd-cafile=/etc/kubernetes/ssl/k8s-root-ca.pem \
 -etcd-certfile=/etc/kubernetes/ssl/kubernetes.pem \
 -etcd-keyfile=/etc/kubernetes/ssl/kubernetes-key.pem \
 -etcd-endpoints=https://192.168.122.11:2379,https://192.168.122.12:2379,https://192.168.122.13:2379 \
 -etcd-prefix=/kubernetes/network \
 -iface=192.168.122.12 \
 -ip-masq
ExecStartPost=/usr/local/bin/mk-docker-opts.sh -k DOCKER_NETWORK_OPTIONS -d /run/flannel/docker
Restart=always
RestartSec=5
StartLimitInterval=0

[Install]
WantedBy=multi-user.target
RequiredBy=docker.service
```



## 安装multus-cni

去multus-cni下载包：
 
```
# wget https://github.com/k8snetworkplumbingwg/multus-cni/releases/download/v3.8/multus-cni_3.8_linux_amd64.tar.gz
```

把解压出来的二进制文件拷贝到所有Kubernetes的worker节点的/opt/cni/bin目录

```
# tar -zxvf multus-cni_3.8_linux_amd64.tar.gz
# mv multus-cni_3.8_linux_amd64/multus-cni /opt/cni/bin/
```


# 配置与创建Pod
## 创建CNI配置
为了保证环境的清洁，请执行如下命令删除worker上的任何已有的cni配置

```
# rm -rf /etc/cni/net.d/*
```

multus使用"delegates"的概念将多个CNI插件组合起来，并且指定一个masterplugin来作为POD的主网络并且被Kubernetes所感知。

然后创建/etc/cni/net.d/10-multus.conf

```
# cat /etc/cni/net.d/10-multus.conf
    {
      "cniVersion": "0.3.1",
      "name": "multus-demo",
      "type": "multus-cni",
      "delegates": [
        {
          "name": "k8snet1",
          "type": "flannel",
          "masterplugin": true,
          "delegate": {
             "isDefaultGateway": true,
             "bridge": "docker0",
             "mtu": 1400
          }
        },
        {
          "name": "k8snet2",
          "type": "macvlan",
          "master": "eth0",
          "ipam": {
                "name": "kube-subnet",
                "type": "kube-ipam",
                "kubeConfig": "/etc/kubernetes/ssl/kube.kubeconfig",
                "etcdConfig": {
                        "etcdURL": "https://192.168.122.11:2379",
                        "etcdCertFile": "/etc/kubernetes/ssl/kubernetes.pem",
                        "etcdKeyFile": "/etc/kubernetes/ssl/kubernetes-key.pem",
                        "etcdTrustedCAFileFile": "/etc/kubernetes/ssl/k8s-root-ca.pem"
                },
                "subnet": "10.188.0.0/16",
                "rangeStart": "10.188.0.10",
                "rangeEnd": "10.188.0.200",
                "gateway": "10.188.0.1",
                "routes": [{
                        "dst": "0.0.0.0/0"
                }],
                "resolvConf": "/etc/resolv.conf"
          }
        }
      ]
    }
```


重启kubelet服务：

```
# systemctl restart kubelet
```


## 创建Database Pod

配置需要固定IP的数据容器

```
# cat db.yaml 
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database
spec:
  selector:
    matchLabels:
      app: database
  serviceName: database
  template:
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database-1
spec:
  selector:
    matchLabels:
      app: database-1
  serviceName: database-1
  template:
    metadata:
      labels:
        app: database-1
      annotations:
        kube-ipam.ip: "10.188.0.218"
        kube-ipam.netmask: "255.255.0.0"
        kube-ipam.gateway: "10.188.0.1"
    spec:
      terminationGracePeriodSeconds: 10
      containers:
      - name: database-1c
        image: 192.168.122.12:5000/db:v1.0
        resources: {}

---

apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database-2
spec:
  selector:
    matchLabels:
      app: database-2
  serviceName: database-2
  template:
    metadata:
      labels:
        app: database-2
      annotations:
        kube-ipam.ip: "10.188.0.219"
        kube-ipam.netmask: "255.255.0.0"
        kube-ipam.gateway: "10.188.0.1"
    spec:
      terminationGracePeriodSeconds: 10
      containers:
      - name: database-2c
        image: 192.168.122.12:5000/db:v1.0
        resources: {} 
```

使用`kubectl apply`命令创建：

```
# 
# kubectl apply -f db.yaml 
statefulset.apps/database created
# 
# kubectl get pod -o wide                
NAME                   READY   STATUS    RESTARTS   AGE     IP            NODE           
database-1-0           1/1     Running   0          2m5s    10.244.69.7   192.168.122.13 
database-2-0           1/1     Running   0          2m5s    10.244.5.5    192.168.122.12 
web-5fd8684df7-8c7zb   1/1     Running   0          3h17m   10.244.71.8   192.168.122.14 
web-5fd8684df7-p9g8s   1/1     Running   0          3h17m   10.244.71.9   192.168.122.14 
#
```



## 创建Web Pod

配置不需要固定IP的web应用


```
# cat web.yaml 
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: web
  name: web
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
  strategy: {}
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - image: 192.168.122.12:5000/nginx:1.7.9
        imagePullPolicy: IfNotPresent
        name: web-c
        ports:
        - containerPort: 80
          name: web
        resources: {}
```

使用`kubectl apply`命令创建：


```
# kubectl apply -f web.yaml               
deployment.apps/web created
# 
# 
# kubectl get pod -o wide 
NAME                   READY   STATUS    RESTARTS   AGE   IP            NODE            
web-5fd8684df7-8c7zb   1/1     Running   0          5s    10.244.71.8   192.168.122.14  
web-5fd8684df7-p9g8s   1/1     Running   0          5s    10.244.71.9   192.168.122.14  
#
```


## 创建service或ingress

用户可以通过web区域网络，通过ingress或service来访问到web服务

给web Pod配置service：

```
# cat web-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: web-svc
  labels:
    app: web
spec:
  type: NodePort
  selector:
    app: web
  ports:
  - port: 80
protocol: TCP

```

使用`kubectl apply`命令创建：

```
#
# kubectl apply -f web-svc.yaml 
service/web-svc created
# 
#
# kubectl get service
NAME         TYPE        CLUSTER-IP      EXTERNAL-IP  PORT(S)       AGE
kubernetes   ClusterIP   10.254.0.1      <none>       443/TCP       7d22h
web-svc      ClusterIP   10.254.150.15   <none>       80:18370/TCP  6m4s
# 
#
```

配置ingress到web service：

```
# 
# cat web-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web-ingress
spec:
  defaultBackend:
    service:
      name: web-svc
      port:
        number: 80
```

使用`kubectl apply`命令创建：

```
# 
[root@localhost ~]# k apply -f web-ingress.yaml 
ingress.networking.k8s.io/web-ingress created
#
```


# 验证分层网络访问

此时，用户可以通过ingress或service来访问到web服务。web pod可以通过database区域网络，访问固定IP地址的database服务。Database区域网络的database Pod可以互相通过固定IP地址进行集群的通信操作。

## 用户通过service访问Web

通过ingress或service来访问到web服务


```
#
# curl http://192.168.122.12:18370
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
    body {
        width: 35em;
        margin: 0 auto;
        font-family: Tahoma, Verdana, Arial, sans-serif;
    }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
#
```



## Web Pod通过固定IP访问Database

查看database-2 Pod的net1网卡，10.188.0.219为固定IP地址

```
#
# kubectl exec -it database-2-0 -- ip address                 
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if9: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1400 qdisc noqueue state UP 
    link/ether 26:16:04:dc:82:fe brd ff:ff:ff:ff:ff:ff
    inet 10.244.69.7/24 brd 10.244.69.255 scope global eth0
       valid_lft forever preferred_lft forever
4: net1@if2: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP 
    link/ether c6:17:6c:da:ee:e9 brd ff:ff:ff:ff:ff:ff
    inet 10.188.0.219/16 brd 10.188.255.255 scope global net1
       valid_lft forever preferred_lft forever
#
```

Web Pod可以通过database区域网络，访问固定IP地址的database服务

```
#
# kubectl exec -it web-5fd8684df7-8c7zb -- ping 10.188.0.219
PING 10.188.0.219 (10.188.0.219): 48 data bytes
56 bytes from 10.188.0.219: icmp_seq=0 ttl=64 time=0.720 ms
56 bytes from 10.188.0.219: icmp_seq=1 ttl=64 time=0.341 ms
56 bytes from 10.188.0.219: icmp_seq=2 ttl=64 time=0.485 ms
56 bytes from 10.188.0.219: icmp_seq=3 ttl=64 time=0.389 ms
56 bytes from 10.188.0.219: icmp_seq=4 ttl=64 time=0.454 ms
^C--- 10.188.0.219 ping statistics ---
5 packets transmitted, 5 packets received, 0% packet loss
round-trip min/avg/max/stddev = 0.341/0.478/0.720/0.131 ms
```


## database区域的Pod通过固定IP互访

database区域网络内的database-1与database-2都拥有固定IP地址，两个数据库Pod之间可以互相通过固定IP进行集群的通信操作。例如，mysql主从架构的时候，主database-1与从database-2之间的同步。

```
#
# kubectl exec -it database-1-0 -- ip address                  
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if9: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1400 qdisc noqueue state UP 
    link/ether 2a:64:f3:b8:18:01 brd ff:ff:ff:ff:ff:ff
    inet 10.244.5.5/24 brd 10.244.5.255 scope global eth0
       valid_lft forever preferred_lft forever
4: net1@if2: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP 
    link/ether fe:9d:18:25:9c:a1 brd ff:ff:ff:ff:ff:ff
    inet 10.188.0.218/16 brd 10.188.255.255 scope global net1
       valid_lft forever preferred_lft forever
```

使用database-2 Pod访问database-1 Pod

```
#
# kubectl exec -it database-2-0 -- /bin/bash
root@database-2-0:/# 
root@database-2-0:/# ip address
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if10: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1400 qdisc noqueue state UP 
    link/ether ba:32:1c:33:dc:a6 brd ff:ff:ff:ff:ff:ff
    inet 10.244.69.7/24 brd 10.244.69.255 scope global eth0
       valid_lft forever preferred_lft forever
4: net1@if2: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP 
    link/ether 86:3b:09:c8:31:93 brd ff:ff:ff:ff:ff:ff
    inet 10.188.0.219/16 brd 10.188.255.255 scope global net1
       valid_lft forever preferred_lft forever
root@database-2-0:/# 
root@database-2-0:/# 
root@database-2-0:/# ping 10.188.0.218
PING 10.188.0.218 (10.188.0.218): 48 data bytes
56 bytes from 10.188.0.218: icmp_seq=0 ttl=64 time=0.335 ms
56 bytes from 10.188.0.218: icmp_seq=1 ttl=64 time=0.246 ms
56 bytes from 10.188.0.218: icmp_seq=2 ttl=64 time=0.484 ms
56 bytes from 10.188.0.218: icmp_seq=3 ttl=64 time=0.371 ms
^C--- 10.188.0.218 ping statistics ---
4 packets transmitted, 4 packets received, 0% packet loss
round-trip min/avg/max/stddev = 0.246/0.359/0.484/0.085 ms
```



# 验证固定IP通信

上文中使用kube-ipam进行固定IP的容器在删除、漂移、重启之后，重新起来的容器依然保持原有的IP地址固定不变。

## 删除一个database Pod

```
#
# kubectl delete pod database-2-0
pod "database-2-0" deleted
#
# kubectl get pod -o wide        
NAME                   READY   STATUS              RESTARTS   AGE     IP            NODE             
database-1-0           1/1     Running             0          20m     10.244.69.7   192.168.122.13   
database-2-0           0/1     ContainerCreating   0          1s      <none>        192.168.122.12   
web-5fd8684df7-8c7zb   1/1     Running             0          3h35m   10.244.71.8   192.168.122.14   
web-5fd8684df7-p9g8s   1/1     Running             0          3h35m   10.244.71.9   192.168.122.14   
# 
```

## 重新启动Pod的IP地址不变


```
#
# kubectl get pod -o wide 
NAME                   READY   STATUS    RESTARTS   AGE     IP            NODE             
database-1-0           1/1     Running   0          20m     10.244.69.7   192.168.122.13   
database-2-0           1/1     Running   0          4s      10.244.5.6    192.168.122.12   
web-5fd8684df7-8c7zb   1/1     Running   0          3h35m   10.244.71.8   192.168.122.14   
web-5fd8684df7-p9g8s   1/1     Running   0          3h35m   10.244.71.9   192.168.122.14   
#
```

查看重新启动Pod的IP地址，我们发现net1网卡的IP地址依然为10.188.0.219

```
#
# kubectl exec -it database-2-0 -- ip address
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if10: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1400 qdisc noqueue state UP 
    link/ether fa:0c:af:f7:ca:4d brd ff:ff:ff:ff:ff:ff
    inet 10.244.5.6/24 brd 10.244.5.255 scope global eth0
       valid_lft forever preferred_lft forever
4: net1@if2: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue state UP 
    link/ether 72:17:7d:5d:fd:fc brd ff:ff:ff:ff:ff:ff
    inet 10.188.0.219/16 brd 10.188.255.255 scope global net1
       valid_lft forever preferred_lft forever
#
```

## 验证容器可以正常访问

使用Web Pod或其他Database Pod访问这个刚刚删除重建的database-2 Pod，可以正常访问：

```
#
# kubectl exec -it web-5fd8684df7-8c7zb -- ping 10.188.0.219
PING 10.188.0.219 (10.188.0.219): 48 data bytes
56 bytes from 10.188.0.219: icmp_seq=0 ttl=64 time=0.720 ms
56 bytes from 10.188.0.219: icmp_seq=1 ttl=64 time=0.341 ms
56 bytes from 10.188.0.219: icmp_seq=2 ttl=64 time=0.485 ms
56 bytes from 10.188.0.219: icmp_seq=3 ttl=64 time=0.389 ms
56 bytes from 10.188.0.219: icmp_seq=4 ttl=64 time=0.454 ms
^C--- 10.188.0.219 ping statistics ---
5 packets transmitted, 5 packets received, 0% packet loss
round-trip min/avg/max/stddev = 0.341/0.478/0.720/0.131 ms
#
```






