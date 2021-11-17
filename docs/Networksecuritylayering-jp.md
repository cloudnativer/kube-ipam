# kube-ipamとmultusに基づいてWebとデータベース階層ネットワークセキュリティアクセスアーキテクチャを実現する。

<br>
<br>

# [1] 基本概況

<br>

## 1.1 kube-ipamとmultusの概要

Kube-ipamは、kubernetesクラスタ内のPod固定IPアドレスに対応しています。いくつかのシーンはIPアドレスに依存していますが、固定IPアドレスのPodを使って、クビ-ipamを使って簡単に解決できます。例えば、mysql主従構造の時、主databaseとdatabaseの間の同期；例えば、keepalivedがクラスタHAをするとき、2つのノード間で通信などを検出する。例えば、いくつかのセキュリティ保護装置は、IPアドレスに基づいてネットワークセキュリティアクセスポリシーの制限を行うシーンなどが必要である。
<br>
multus-CNI同時に複数のネットワークインターフェースをkubernetes環境内のPodに同時に追加することをサポートしています。このような配置は、アプリケーションネットワークやデータベースなど複数のネットワーク領域を相互に分離し、コンテナクラスターネットワークアーキテクチャを効果的に制御するために、セキュリティ要員が有利である。

<br>

## 1.2 ネットワーク階層構造設計
 
<br>

![kube-ipam](images/Networksecuritylayering.jpg)

<br>

上の図は、各Podが2つのインターフェースを有することを示している。eth0、net1。eth0は外部ユーザとしてweb podのネットワークインターフェースにアクセスする。一方、net1は付加的なコンテナネットワークカードであり、web Podとしてdatabase Podまでの内部ネットワーク通信である。

<br>
<br>


# [2] CNIプラグインのインストール

<br>

## 2.1 cni pluginとflannelをインストールします

<br>

### cni pluginの設置

```
# wget https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz
# tar -zxvf cni-plugins-linux-amd64-v0.9.1.tgz -C /opt/cni/bin/
```

### flanneldの設置

flanneldを作成するために必要なsubnetセグメント

```
# etcdctl --endpoints=https://192.168.1.11:2379,https://192.168.1.12:2379,https://192.168.1.13:2379  --ca-file=/etc/kubernetes/ssl/k8s-root-ca.pem --cert-file=/etc/kubernetes/ssl/kubernetes.pem --key-file=/etc/kubernetes/ssl/kubernetes-key.pem set /kubernetes/network/config '{"Network":"10.244.0.0/16", "SubnetLen":24, "Backend":{"Type":"vxlan"}}'
```

flanneldのパッケージをダウンロードします：

```
# wget https://github.com/flannel-io/flannel/releases/download/v0.11.0/flannel-v0.11.0-linux-amd64.tar.gz
# tar -zxvf flannel-v0.11.0-linux-amd64.tar.gz -C /opt/cni/bin/
```

編集 /etc/systemd/system/flanneld.service

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

修正/etc/systemd/system/docker.service、Evironment Fileを追加し、ExecStartパラメータを変更することに注意してください：

```
...
[Service]
EnvironmentFile=/run/flannel/docker
ExecStart=/usr/bin/dockerd $DOCKER_NETWORK_OPTIONS
...
```

<br>

説明：もしあなたが使用しているのが<a href="https://github.com/cloudnativer/kube-install">kube-innstall</a>にインストールされたkubernetesクラスタは、cni pluginとflannelをインストールする2ステップを省略することができます。

<br>

## 2.2 kube-ipamの設置

<a href="docs/download.md">ダウンロード</a>または<a href="docs/build.md">コンパイル</a>を通じて、`kube-i pam`のバイナリファイルを取得して、k 8 s-nodeホストの`/opt/cni/bin/ディレクトリにコピーしてください。

```
# wget https://github.com/cloudnativer/kube-ipam/releases/download/v0.2.0/kube-ipam-v0.2.0-x86.tgz
# tar -zxvf kube-ipam-v0.2.0-x86.tgz
# mv kube-ipam-v0.2.0-x86/kube-ipam /opt/cni/bin/kube-ipam
```


<br>


## 2.3 multus-cniの設置

multus-cniパッケージをダウンロード：
 
```
# wget https://github.com/k8snetworkplumbingwg/multus-cni/releases/download/v3.8/multus-cni_3.8_linux_amd64.tar.gz
```

解凍されたバイナリファイルをすべてのKubernetesのワーカーノードの/opt/cni/binディレクトリにコピーします。

```
# tar -zxvf multus-cni_3.8_linux_amd64.tar.gz
# mv multus-cni_3.8_linux_amd64/multus-cni /opt/cni/bin/
```

<br>
<br>

# [3] Podの設定と作成

<br>

# 3.1 CNI構成を作成する

<br>

ホスト環境の清潔を確保するために、k8s-nodeホスト上の既存のcniの削除を以下のコマンドで実行してください。

```
# rm -rf /etc/cni/net.d/*
```

multusは「delegates」の概念を使用して複数のCNIプラグインを組み合わせ、PODのマスターネットワークとしてmasterpluginを指定し、Kubernetesによって知覚される。

それから/etc/cni/net.d/10-multus.confを作成します。

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
                        "etcdURL": "https://192.168.1.50:2379,https://192.168.1.58:2379,https://192.168.1.63:2379",
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

kubeletサービスを再起動する：

```
# systemctl restart kubelet
```

<br>

## 3.2 Database Podを作成します

IPを固定したいデータコンテナを設定します。

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

`kubectl apply`コマンドを使って作成します：

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

## 3.3 Web Podを作成

固定IPが不要なwebアプリケーションを設定します：

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

`kubectl apply`コマンドを使って作成します：

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

## 3.4 serviceまたはingressを作成します

ユーザーはwebエリアネットワークを通じて、ingressまたはserviceを通じてウェブサービスにアクセスできます。

web Podにserviceを設定する：

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

`kubectl apply`コマンドを使って作成します：

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

ingressをウェブserviceに設定します：

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

`kubectl apply`コマンドを使って作成します：

```
# 
[root@localhost ~]# k apply -f web-ingress.yaml 
ingress.networking.k8s.io/web-ingress created
#
```

<br>
<br>

# [4] 階層ネットワークアクセスの検証

<br>

このとき、ユーザは、ingressまたはserviceを介してウェブサービスにアクセスすることができる。web podは、databaseエリアネットワークを通じて、固定IPアドレスのdatabaseサービスにアクセスすることができます。Database領域ネットワークのdatabase Podは、互いに固定IPアドレスを介してクラスタの通信動作を行うことができる。

<br>

## 4.1 ユーザーはserviceを通じてWebにアクセスします

ingressまたはserviceを通じてウェブサービスにアクセスします。

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

## 4.2 Web Podは固定IPでDatabaseにアクセスする

database-2 Podのnet1ネットカードを調べます。10.188.0.219は固定IPアドレスです。

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

Web Podは、databaseエリアネットワークを通じて、固定IPアドレスのdatabaseサービスにアクセスできます

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

## 4.3 databaseエリアのPodは固定IPで相互訪問します

database領域ネットワーク内のdatabase-1とdatabase-2は固定IPアドレスを有しており、2つのデータベースPodの間で互いに固定IPを介してクラスタの通信動作が可能である。例えば、mysql主従アーキテクチャの場合、主database-1とdatabase-2の間の同期。

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

database-2 Podを使ってdatabase-1 Podにアクセスします

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

# [5] 固定IP通信の検証

<br>

上記では`kube-ipam`を使用して固定IPを行う容器は、削除、ドリフト、再起動後も、元のIPアドレスが固定されています。

<br>

## 5.1 database Podを削除します

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

## 5.2 Podを再起動するIPアドレスは不変です


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

PodのIPアドレスを再起動します。net1ネットカードのIPアドレスは依然として10.188.0.219です

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

## 5.3 検証容器は正常にアクセスできます

Web Podまたは他のDatabase Podを使用して、この削除されたばかりの再構築されたDatabase-2 Podにアクセスして、正常にアクセスできます：

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


