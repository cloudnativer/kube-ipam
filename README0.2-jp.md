Kube-ipamは、etcd分布式記憶に基づいて、クラスタ内のIPアドレスの一意性を確保するために、kubernetes動的IPネットワーク割当管理を実現する。Kub-ipamは、kubernetesクラスタ内のPod固定IPアドレスをサポートし、同時に、resolv.confのDNS構成をサポートする。

<br>

![kube-ipam](docs/images/kube-ipam-logo.jpg)

<br>

言語を切り替え: <a href="README0.2.md">English Documents</a> | <a href="README0.2-zh-hk.md">繁體中文檔案</a> | <a href="README0.2-zh.md">简体中文文档</a> | <a href="README0.2-jp.md">日本語の文書</a>

<br>
<br>

# [1] 概要

いくつかのシーンではIPアドレスに依存していますが、固定IPアドレスのPodを使うと、`Kube-ipam`を使って簡単にこのような問題を解決できます。例えば、mysql主従構造の時、主databaseとdatabaseの間の同期；例えば、keepalivedがクラスタHAをするとき、2つのノード間で通信などを検出する。例えば、いくつかのセキュリティ保護装置は、IPアドレスに基づいてネットワークセキュリティアクセスポリシーの制限を行うシーンなどが必要です。

<br>

![kube-ipam](docs/images/kube-ipam02.jpg)

<br>

`Kube-ipam`はetcd分布式記憶に基づいて、kubernetes動的IPネットワーク割り当て管理を実現し、kubernetesクラスタ内のPodが固定的なIPアドレスを持つことを確保する。kube-ipam配置を使用した後、上図のfixed-in Podは破壊再構築後も元のIPアドレスの固定を維持することができます。


<br>
<br>

# [2] kube-ipamの取り付け

<a href="docs/download.md">ダウンロード</a>または<a href="docs/build.md">コンパイル</a>を通じて、`kube-i pam`のバイナリファイルを取得して、k 8 s-nodeホストの`/opt/cni/bin/ディレクトリにコピーしてください。


```
# wget https://github.com/cloudnativer/kube-ipam/releases/download/v0.2.0/kube-ipam-v0.2.0-x86.tgz
# tar -zxvf kube-ipam-v0.2.0-x86.tgz
# mv kube-ipam-v0.2.0-x86/kube-ipam /opt/cni/bin/kube-ipam
```

<br>
<br>

# [3] /etc/cni/net.d配置

## 3.1 サブネットとetcd配置

あなたは`subnet`パラメータでIPサブネットの情報を設定し、`gateway`を通じてゲートウェイの情報を設定することができます。etcdConfig`でETcdの証明書とendpointアドレスを設定できます。
すべてのk8s-nodeホストの`/etc/cni/net.d/1-kube-ipam.conf`ファイルを編集します。

```
# cat /etc/cni/net.d/1-kube-ipam.conf
{
        "cniVersion":"0.3.1",
        "name": "k8snetwork",
        "type": "macvlan",
        "master": "eth1",
        "ipam": {
                "name": "kube-subnet",
                "type": "kube-ipam",
		"kubeConfig": "/etc/kubernetes/pki/kubectl.kubeconfig"
                "etcdConfig": {
                        "etcdURL": "https://192.168.1.50:2379,https://192.168.1.58:2379,https://192.168.1.63:2379",
                        "etcdCertFile": "/etc/kubernetes/pki/etcd.pem",
                        "etcdKeyFile": "/etc/kubernetes/pki/etcd-key.pem",
                        "etcdTrustedCAFileFile": "/etc/kubernetes/pki/ca.pem"
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

```

## 3.2  配置パラメータの説明

* `type` (string, required): CNIプラグインの種類を記入します。例えば、macvlan、ipvlan、kube-router、bridgeなどです（`Multius`と組み合わせてより多くのCNIプラグインをサポートすることもできます）。
* `routes` (string, optional): コンテナの名前空間のルートリストに追加します。各ルーティングは、`dst`とオプションの`gw`フィールドを有するものである。`gw`を省略すると、「ゲートウェイ」の値が使用されます。
* `resolvConf` (string, optional): ホスト上で解析され、DNS構成として返される`resov.co nf`ファイルパス。
* `ranges`, (array, required, nonempty) an array of arrays of range objects:
	* `subnet` (string, required): 割り当てられたCIDRブロックです。
	* `rangeStart` (string, optional): `subnet`サブネットから配信されるIPアドレスは、デフォルトでは`subnet`サブネット内の「.2」というIPアドレスです。
	* `rangeEnd` (string, optional): `subnet`子ネットの中から分配のIPアドレスを終わって、デフォルトは`subnet`子ネットの中の“.254”のこのIPアドレスです。
	* `gateway` (string, optional): `subnet`サブネットから割り当てられたゲートウェイのIPアドレスは、デフォルトでは`subnet`サブネット内の「.1」というIPアドレスです。
* `etcdConfig`：etcdアドレス情報の対象
  * `etcdURL` (string, required): etcdのendpoint URLアドレスです。
  * `etcdCertFile` (string, required): etcdのcertファイル。
  * `etcdKeyFile` (string, required): etcdのkeyファイル。
  * `etcdTrustedCAFileFile` (string, required): etcdのcaファイル。

<br>
<br>

# [4] Kubenetes固定IP容器方法

## 4.1 固定IPアドレス構成

pod IPアドレスの固定割り当ては、podの`annotations`に`kube-ipam.ip`を配置し、`kube-ipam.netmark`と`kube-ipam.gateway`パラメータを配置することで実現できます。
<br>
`/etc/cni/net.d/1-kube-ipam.com.com`において、ランダムIPアドレスの範囲は`rangestart`と`rangeend`に設定されている。`rangestart`と`rangeend`のIPアドレスセグメントには設定されていません。固定IPの容器に手動で割り当てられます。
<br>
言い換えれば、podのIPアドレスを固定したいなら、`kube-i pam.ip`の値を`rangestart`と`rangeend`の範囲に設定しないでください。
<br>
新しい`fixed-i-p-test-Deployment.yaml`を作成して、固定IPのPodを作成します。

```
# cat fixed-ip-test-Deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fixed-ip-test
  namespace: default
  labels:
    k8s-app: cloudnativer-test
spec:
  replicas: 1
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  selector:
    matchLabels:
      k8s-app: cloudnativer-test
  template:
    metadata:
      labels:
        k8s-app: cloudnativer-test
      annotations:
        kube-ipam.ip: "10.188.0.216"
        kube-ipam.netmask: "255.255.0.0"
        kube-ipam.gateway: "10.188.0.1"
    spec:
      containers:
      - name: fixed-ip-test
        image: nginx:1.7.9
        imagePullPolicy: IfNotPresent
        ports:
        - name: http
          containerPort: 80  
---

```

この例では、10.188.00/16のセグメントで、10.188.0.10～10.188.0.200以外のIPアドレスをPodに割り当てることができます。
<br>

説明：ランダムIPのPodを作成するには、annotationsの`kube-i pam.ip`、`kube-ipam.netmark`と`kube-ipam.gateway`を削除して配置すればいいです。

<br>

# 4.2 固定IPのPodを作成する。

`kubectlアプリ-f`コマンドを使って固定IPのPodを作成します。

```
# kubectl apply -f fixed-ip-test-Deployment.yaml
#
# kubectl get pod -o wide
  NAME                             READY   STATUS    RESTARTS   AGE     IP             NODE   
  fixed-ip-test-6d9b74fd4d-dbbsd   1/1     Running   0          2d23h   10.188.0.216   192.168.20.21

```

現在、このfixed-inp-test-6 d 9 b 74 fd 4 d-dbbbbbsというPodは固定不変のIPアドレスを割り当てられました（10.188.0.216）。

# 4.3 Podを再構築し、IPは固定されたままであること。

例えば、上記のPodを削除するために`kubectl delete`コマンドを使用します。kubernetesは新しいPodを自動的に再構築します。

```
# kubectl delete pod fixed-ip-test-6d9b74fd4d-dbbsd
#
# kubectl get pod -o wide
  NAME                             READY   STATUS    RESTARTS   AGE   IP             NODE   
  fixed-ip-test-6d9b74fd4d-xjhek   1/1     Running   0          1h    10.188.0.216   192.168.30.35

```

この時、新たに起動されたfixed-ip-test-6d9b74fd4d-xjhekというPodのIPアドレスは依然として10.188.0.216です。

<br>
<br>

# [5] ログ情報を確認する。

k8s-nodeホスト上の`/var/log/kube-ipam.log`ファイルを見て、`kube-ipam`のシステムログ情報を得ることができます。

<br>
<br>

# [6] 階層ネットワークセキュリティアーキテクチャ

<br>
`kube-ipam`は`Multius`と結合してネットワークを組むことができ、より多くのCNIプラグインシーンでのコンテナIPアドレスの固定をサポートすることができます。例えば、私たちは、`kube-ipam`と`Multius`に基づいて、Webとデータベースの階層ネットワークセキュリティアクセスアーキテクチャを実現し、Podが同時にランダムIPと固定IPなどの複数のネットワークインターフェースをサポートするようにすることができる。このような配置は、アプリケーションネットワークやデータベースなど複数のネットワーク領域を相互に分離し、コンテナクラスターネットワークアーキテクチャを効果的に制御するために、セキュリティ要員が有利である。

<br>

![kube-ipam](docs/images/Networksecuritylayering.jpg)

<br>

上の図は、各Podが2つのインターフェースを有することを示している eth 0、net 1。eth 0は外部ユーザとしてweb podのネットワークインターフェースにアクセスする。一方、net 1は付加的なコンテナネットワークカードであり、web Podとしてdatabase Podまでの内部ネットワーク通信である。

<br/>

ユーザは、ingressまたはserviceを介してウェブサービスにアクセスすることができる。web podは、databaseエリアネットワークを通じて、固定IPアドレスのdatabaseサービスにアクセスすることができます。Database領域ネットワークのdatabase Podは、互いに固定IPアドレスを介してクラスタの通信動作を行うことができる。階層ネットワークセキュリティアクセスアーキテクチャの<a href="docs/Networksecuritylayering-jp.md">インストールと配置はここをクリックしてください</a>。

<br>
<br>

# [7] IssuesとPRの提出を歓迎します。

使用中に問題があったら，をクリックしてもいいです<a href="https://github.com/cloudnativer/kube-ipam/issues">https://github.com/cloudnativer/kube-install/issues</a>Issuesを提出してもいいです。Forkソースコードを修正してBUGを修復してみて、PRを提出してください。<br>

```
# git clone your-fork-code
# git checkout -b your-new-branch
# git commit -am "Fix bug or add some feature"
# git push origin your-new-branch
```
<br>
IssuesとPRを提出してください。
<br>
貢献者の皆様、ありがとうございます。

<br>
<br>
<br>

