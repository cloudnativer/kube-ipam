# Implementation of hierarchical network security access architecture for Web and database based on kube-ipam and Multus

<br>
<br>

# Introduction

<br>

## Overview of kube-ipam and Multus
Kube-ipam supports fixed IP addresses for Pods in the kubernetes cluster. Some scenarios often rely on IP addresses and need to use Pods with fixed IP addresses. You can use kube-ipam to easily solve this type of problem. For example, in the mysql master-slave architecture, the synchronization between the master database and the slave database; for example, when keepalived is doing cluster HA, the communication between the two nodes is detected; for example, some security protection equipment needs to be based on IP addresses for network security Scenarios restricted by access policies, etc.
<br>
Multus-CNI supports adding multiple network interfaces to Pods in the kubernetes environment at the same time. This deployment method is conducive to security personnel to isolate multiple network areas such as application networks and databases from each other, and effectively control the container cluster network architecture.

<br>

## Network layered architecture design
 
<br>

![kube-ipam](images/Networksecuritylayering.jpg)

<br>

The figure above shows that each Pod has 2 interfaces: eth0 and net1. Eth0 serves as the network interface for external users to access the web pod; and net1 is an additional container network card that serves as the internal network communication from the web Pod to the database Pod.

<br>
<br>


# Install CNI plugin

<br>

## Install cni plugin and flannel

<br>

### Install cni plugin

```
# wget https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz
# tar -zxvf cni-plugins-linux-amd64-v0.9.1.tgz -C /opt/cni/bin/
```

### Install flanneld

Create the subnet network segment required by flanneld

```
# etcdctl --endpoints=https://192.168.1.11:2379,https://192.168.1.12:2379,https://192.168.1.13:2379  --ca-file=/etc/kubernetes/ssl/k8s-root-ca.pem --cert-file=/etc/kubernetes/ssl/kubernetes.pem --key-file=/etc/kubernetes/ssl/kubernetes-key.pem set /kubernetes/network/config '{"Network":"10.244.0.0/16", "SubnetLen":24, "Backend":{"Type":"vxlan"}}'
# wget https://github.com/flannel-io/flannel/releases/download/v0.11.0/flannel-v0.11.0-linux-amd64.tar.gz
# tar -zxvf flannel-v0.11.0-linux-amd64.tar.gz -C /opt/cni/bin/
```

Edit /etc/systemd/system/flanneld.service

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
 -etcd-endpoints=https://192.168.1.11:2379,https://192.168.1.12:2379,https://192.168.1.13:2379 \
 -etcd-prefix=/kubernetes/network \
 -ip-masq
ExecStartPost=/usr/local/bin/mk-docker-opts.sh -k DOCKER_NETWORK_OPTIONS -d /run/flannel/docker
Restart=always
RestartSec=5
StartLimitInterval=0

[Install]
WantedBy=multi-user.target
RequiredBy=docker.service
```

<br>

Notice: If you are using kubernetes cluster installed by <a href="https://github.com/cloudnativer/kube-install">kube-install</a>, then these two steps can be omitted。

<br>

## Install kube-ipam

kube-ipam binary program files can be obtained by <a href="docs/download.md">download</a> or <a href="docs/build.md">compile</a>, and copy the kube-ipam binary to the `/opt/cni/bin/` directory
```
# tar -zxvf kube-ipam-x86.tgz
# mv kube-ipam /opt/cni/bin/kube-ipam
```

<br>


## Install multus-cni

Download the multus-cni package:
 
```
# wget https://github.com/k8snetworkplumbingwg/multus-cni/releases/download/v3.8/multus-cni_3.8_linux_amd64.tar.gz
```

Copy the decompressed binary file to the /opt/cni/bin directory of all Kubernetes worker nodes

```
# tar -zxvf multus-cni_3.8_linux_amd64.tar.gz
# mv multus-cni_3.8_linux_amd64/multus-cni /opt/cni/bin/
```

<br>
<br>

# Configure and create Pod

<br>

## Create CNI configuration

<br>
In order to ensure a clean host environment, please execute the following command to delete the existing cni configuration on the kubernetes node host:

```
# rm -rf /etc/cni/net.d/*
```

Multus uses the concept of "delegates" to combine multiple CNI plugins, and designates a masterplugin as the main network of the POD and is perceived by Kubernetes.

Then create /etc/cni/net.d/10-multus.conf

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
                        "etcdURL": "https://192.168.1.11:2379",
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


Restart the kubelet service:

```
# systemctl restart kubelet
```

<br>

## Create Database Pod

Configure data containers that require a fixed IP: 

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
        image: 192.168.1.12:5000/db:v1.0
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
        image: 192.168.1.12:5000/db:v1.0
        resources: {} 
```

Use the `kubectl apply` command to create:

```
# 
# kubectl apply -f db.yaml 
statefulset.apps/database created
# 
# kubectl get pod -o wide                
NAME                   READY   STATUS    RESTARTS   AGE     IP            NODE           
database-1-0           1/1     Running   0          2m5s    10.244.69.7   192.168.1.13 
database-2-0           1/1     Running   0          2m5s    10.244.5.5    192.168.1.12 
web-5fd8684df7-8c7zb   1/1     Running   0          3h17m   10.244.71.8   192.168.1.14 
web-5fd8684df7-p9g8s   1/1     Running   0          3h17m   10.244.71.9   192.168.1.14 
#
```

<br>

## Create Web Pod

Configure web applications that do not require a fixed IP: 


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
      - image: 192.168.1.12:5000/nginx:1.7.9
        imagePullPolicy: IfNotPresent
        name: web-c
        ports:
        - containerPort: 80
          name: web
        resources: {}
```

Use the `kubectl apply` command to create:


```
# kubectl apply -f web.yaml               
deployment.apps/web created
# 
# 
# kubectl get pod -o wide 
NAME                   READY   STATUS    RESTARTS   AGE   IP            NODE            
web-5fd8684df7-8c7zb   1/1     Running   0          5s    10.244.71.8   192.168.1.14  
web-5fd8684df7-p9g8s   1/1     Running   0          5s    10.244.71.9   192.168.1.14  
#
```

<br>

## Create service or ingress

Users can access web services through ingress or service through the web area network. 

Configure service for web Pod:

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

Use the `kubectl apply` command to create:

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

Configure ingress to web service:

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

Use the `kubectl apply` command to create:

```
# 
[root@localhost ~]# k apply -f web-ingress.yaml 
ingress.networking.k8s.io/web-ingress created
#
```

<br>
<br>

# Verify hierarchical network access

<br>

At this point, the user can access the web service through ingress or service. The web pod can access the database service with a fixed IP address through the database area network. The database Pods of the Database area network can communicate with each other in the cluster through a fixed IP address.
<br>

## The user accesses the Web through the service

Access to web services through ingress or service


```
#
# curl http://192.168.1.12:18370
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

<br>

## Web Pod accesses the database through a fixed IP

Check the net1 network card of database-2 Pod, 10.188.0.219 is the fixed IP address

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

Web Pod can access the database service with fixed IP address through the database area network

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

<br>

## Pods in the database area communicate with each other through a fixed IP

Both database-1 and database-2 in the database area network have fixed IP addresses, and the two database Pods can communicate with each other through a fixed IP cluster. For example, in the mysql master-slave architecture, the synchronization between the master database-1 and the slave database-2.

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

Use database-2 Pod to access database-1 Pod

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

<br>
<br>

# Verify fixed IP communication

<br>

After the container that uses kube-ipam to fix the IP above is deleted, drifted, and restarted, the rebuilt container still keeps the original IP address fixed.

<br>

## Delete a database Pod

```
#
# kubectl delete pod database-2-0
pod "database-2-0" deleted
#
# kubectl get pod -o wide        
NAME                   READY   STATUS              RESTARTS   AGE     IP            NODE             
database-1-0           1/1     Running             0          20m     10.244.69.7   192.168.1.13   
database-2-0           0/1     ContainerCreating   0          1s      <none>        192.168.1.12   
web-5fd8684df7-8c7zb   1/1     Running             0          3h35m   10.244.71.8   192.168.1.14   
web-5fd8684df7-p9g8s   1/1     Running             0          3h35m   10.244.71.9   192.168.1.14   
# 
```

<br>

## Restart the Pod's IP address unchanged


```
#
# kubectl get pod -o wide 
NAME                   READY   STATUS    RESTARTS   AGE     IP            NODE             
database-1-0           1/1     Running   0          20m     10.244.69.7   192.168.1.13   
database-2-0           1/1     Running   0          4s      10.244.5.6    192.168.1.12   
web-5fd8684df7-8c7zb   1/1     Running   0          3h35m   10.244.71.8   192.168.1.14   
web-5fd8684df7-p9g8s   1/1     Running   0          3h35m   10.244.71.9   192.168.1.14   
#
```

Checking the IP address of the restarted Pod, we found that the IP address of the net1 network card is still 10.188.0.219.

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

<br>

## Verify that the container can be accessed normally

Use Web Pod or other Database Pod to access this newly deleted and rebuilt database-2 Pod, you can access it normally:

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




