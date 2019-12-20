#### 使用Kubeadm搭建kubernetes的准备工作

基础环境配置

#### 1. 配置静态ip地址

* 修改IP地址为静态ip地址
```bash
>  vi /etc/sysconfig/network-scripts/ifcfg-ens33
BOOTPROTO=static #dhcp改为static（修改）
ONBOOT=yes #开机启用本配置，一般在最后一行（修改）

IPADDR=192.168.179.111 #静态IP（增加）
GATEWAY=192.168.179.2 #默认网关，虚拟机安装的话，通常是2，也就是VMnet8的网关设置（增加）
NETMASK=255.255.255.0 #子网掩码（增加）
DNS1=192.168.179.2 #DNS 配置，虚拟机安装的话，DNS就网关就行，多个DNS网址的话再增加（增加）
```

*  重启生效静态ip地址配置
```bash
> systemctl restart network
```

*  相互 ping对方ip地址
```bash
> ping www.baidu.com 
```

#### 2. 配置节点间无密钥通信

* 三个节点分别执行：
```bash
> ssh-keygen
```
ssh 公钥认证是ssh认证的方式之一。通过公钥认证可实现ssh免密码登陆，git的ssh方式也是通过公钥进行认证的。

在用户目录的home目录下，有一个.ssh的目录，和当前用户ssh配置认证相关的文件，几乎都在这个目录下。

`ssh-keygen` 可用来生成ssh公钥认证所需的公钥和私钥文件。

使用 `ssh-keygen` 时，请先进入到 `~/.ssh` 目录，不存在的话，请先创建。并且保证 `~/.ssh` 以及所有父目录的权限不能大于 711.

使用 `ssh-kengen` 会在`~/.ssh/`目录下生成两个文件，不指定文件名和密钥类型的时候，默认生成的两个文件是：

* id_rsa 私钥文件
* id_rsa.pub 公钥文件


生成ssh key的时候，可以通过 -f 选项指定生成文件的文件名，如下:
```bash
> ssh-keygen -f test   -C "test key"
                --文件名     --备注
```

如果没有指定文件名，会询问你输入文件名:
```bash
> ssh-keygen
  Generating public/private rsa key pair.
  Enter file in which to save the key (/home/huqiu/.ssh/id_rsa):
```
你可以输入你想要的文件名，这里我们输入k8s。

之后，会询问你是否需要输入密码。输入密码之后，以后每次都要输入密码。请根据你的安全需要决定是否需要密码，如果不需要，直接回车:

```bash
>  ssh-keygen -t rsa -f k8s -C "k8s key"
Generating public/private rsa key pair.
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
```

为了让私钥文件和公钥文件能够在认证中起作用，请确保权限正确。

对于`.ssh` 以及父文件夹，当前用户用户一定要有执行权限，其他用户最多只能有执行权限。

对于公钥和私钥文件也是: 当前用户一定要有执行权限，其他用户最多只能有执行权限。

* 三个节点分别执行：
```bash
> ssh-copy-id -i ~/.ssh/id_rsa.pub root@192.168.79.132
```

* 测试：
```bash
> ssh root@192.168.79.132
```


#### 3. 配置hostname

CentOS7永久修改：
```bash
> hostnamectl set-hostname master
```

#### 4. 配置本地DNS解析
```bash
> vim /etc/hosts

192.168.79.130 master
192.168.79.131 node1
192.168.79.132 node2
```


复制到各个节点：
```bash
> scp /etc/hosts root@192.168.79.132:/etc/
```
	
#### 5. 关闭selinux和firewalld
```bash
> systemctl stop firewalld && systemctl disable firewalld

>  sed -i 's/^SELINUX=enforcing$/SELINUX=disabled/' /etc/selinux/config && setenforce 0
```

#### 6. 关闭swap内存交换空间

```bash
> swapoff -a
> yes | cp /etc/fstab /etc/fstab_bak
> cat /etc/fstab_bak |grep -v swap > /etc/fstab
```

#### 7. 设置时间同步

* 设置时区

```bash
> timedatectl set-timezone Asia/Shanghai
> yum install -y chrony
> sed -i 's/^server/#&/' /etc/chrony.conf
```


* 设置上游ntp服务器

```bash
> cat >> /etc/chrony.conf << EOF
server 0.asia.pool.ntp.org iburst
server 1.asia.pool.ntp.org iburst
server 2.asia.pool.ntp.org iburst
server 3.asia.pool.ntp.org iburst
allow all
EOF
```

* 设置为开机自启动
 
```bash
> systemctl enable chronyd && systemctl restart chronyd
```

* 开启网络时间同步

```bash
> timedatectl set-ntp true
``` 


* 开始同步时间
```bash
> chronyc sources
```

#### 8.修改iptables参数

* iptables

纠正iptables被绕过导致流量路由不正确

```bash
> cat <<EOF >  /etc/sysctl.d/k8s.conf
vm.swappiness = 0
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
EOF
```

* 生效配置

```bash
> modprobe br_netfilter
sysctl -p /etc/sysctl.d/k8s.conf
```

加载ipvs相关模块:

* 配置
```bash
cat <<EOF > /etc/sysconfig/modules/ipvs.modules 
#!/bin/bash
modprobe -- ip_vs
modprobe -- ip_vs_rr
modprobe -- ip_vs_wrr
modprobe -- ip_vs_sh
modprobe -- nf_conntrack_ipv4
EOF	
```

* 执行脚本开机生效
```bash
> chmod 755 /etc/sysconfig/modules/ipvs.modules && bash /etc/sysconfig/modules/ipvs.modules && lsmod | grep -e ip_vs -e nf_conntrack_ipv4
```
* 安装工具查看ipvs的代理规则
```bash
> yum install ipset ipvsadm -y
```

#### 9. 配置阿里云docker源

* 安装yum-config-manager工具:

```bash
> yum -y install yum-utils 
```

* 配置源:

```bash
> yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
```

* 安装最新版dockers-ce:

```bash
> yum install -y docker-ce
```

* 设置开机自启动:

```bash
> systemctl start docker && systemctl enable docker
```

可设置daocloud镜像加速或者阿里云镜像加速，（不操作节约时间）.

#### 10. 修改docker默认配置

* 编辑docker.service:

```bash
> vim /usr/lib/systemd/system/docker.service
# 增加配置项：配置docker科学上网代理
Environment="HTTPS_PROXY=http://192.168.43.162:1080"
Environment="NO_PROXY=127.0.0.1/8,192.168.0.0/16"
ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock
ExecStartPost=/usr/sbin/iptables -P FORWARD ACCEPT
ExecReload=/bin/kill -s HUP $MAINPID
```


* 使dockers配置生效:
```bash
> systemctl daemon-reload
> systemctl restart docker

> docker info  #查看配置是否生效

Containers: 21
 Running: 15
 Paused: 0
 Stopped: 6
Images: 22
Server Version: 18.06.2-ce
Storage Driver: overlay2
 Backing Filesystem: extfs
 Supports d_type: true
 Native Overlay Diff: true
Logging Driver: json-file
Cgroup Driver: cgroupfs
Plugins:
 Volume: local
 Network: bridge host macvlan null overlay
 Log: awslogs fluentd gcplogs gelf journald json-file logentries splunk syslog
Swarm: inactive
Runtimes: runc
Default Runtime: runc
Init Binary: docker-init
containerd version: 468a545b9edcd5932818eb9de8e72413e616e86e
runc version: 69663f0bd4b60df09991c08812a60108003fa340
init version: fec3683
Security Options:
 seccomp
  Profile: default
Kernel Version: 3.10.0-862.11.6.el7.x86_64
Operating System: CentOS Linux 7 (Core)
OSType: linux
Architecture: x86_64
CPUs: 4
Total Memory: 15.51GiB
Name: zqg-beta-es-101
ID: O4CB:TCJN:ER57:64RZ:IXNJ:OPRR:WBGJ:YM7W:UU63:CMG6:B3CI:ZSUJ
Docker Root Dir: /var/lib/docker
Debug Mode (client): false
Debug Mode (server): false
HTTPS Proxy: http://192.168.43.162:1080
No Proxy: 127.0.0.1/8,192.168.0.0/16
Registry: https://index.docker.io/v1/
Labels:
Experimental: false
Insecure Registries:
 127.0.0.0/8
Live Restore Enabled: false

```
打印当前bridge变量和值：
```bash
> sysctl -a | grep bridge

net.bridge.bridge-nf-call-arptables = 0
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-filter-pppoe-tagged = 0
net.bridge.bridge-nf-filter-vlan-tagged = 0
net.bridge.bridge-nf-pass-vlan-input-dev = 0
sysctl: reading key "net.ipv6.conf.all.stable_secret"
sysctl: reading key "net.ipv6.conf.cni0.stable_secret"
sysctl: reading key "net.ipv6.conf.default.stable_secret"
sysctl: reading key "net.ipv6.conf.docker0.stable_secret"
sysctl: reading key "net.ipv6.conf.eth0.stable_secret"
sysctl: reading key "net.ipv6.conf.flannel/1.stable_secret"
sysctl: reading key "net.ipv6.conf.lo.stable_secret"
sysctl: reading key "net.ipv6.conf.veth173eab99.stable_secret"
sysctl: reading key "net.ipv6.conf.vethcd0c949.stable_secret"
```

#### 11. 配置阿里云的kubernetes的源

* 配置kubernetes的源
```bash
> vim  /etc/yum.repo.d/kubernetes.repo
[kubernetes]
name=Kubernetes Repository
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64/
gpgcheck=1
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg
```

* 验证是否配置成功
```bash
> yum repolist 
```

* 查看kubernetes软件包
```bash
> yum list all | grep "^kube"
```

#### 12. 安装kubernetes软件包 

在安装kubernetes的包的时候,上面的所有步骤在三个节点均需执行.

安装kubeadm,kubelet,kubectl:

```bash
> yum install kubeadm kubelet kubectl -y 
```

查看安装位置:
```bash
> rpm -ql kubelet
> rpm -ql kubeadm
```

初始化之前的配置，设置忽略swap报错:

```bash
> vi /etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS="--fail-swap-on=false"
```

查看初始化配置选项：
```bash
> 	kubeadm config print init-defaults
  		
flannel:10.244.0.0/16 # 默认网络 
calico:192.168.0.0/16 # 默认网络  	
```
修改默认配置两种方法：

* 使用传递选项 命令行

* 使用配置文件

初始化kubeadm版本:
```bash
> kubeadm init --kubernetes-version=v1.15.1
```

查看kubeadm版本：
```bash
> rpm -q kubeadm
```

指定pod网络 使用flannel:
```bash
> kubeadm init --kubernetes-version="v1.15.1" --pod-network-cidr="10.244.0.0/16" --dry-run
```

查看需要的包:
```bash
> kubeadm config images list

k8s.gcr.io/kube-apiserver:v1.15.2
k8s.gcr.io/kube-controller-manager:v1.15.2
k8s.gcr.io/kube-scheduler:v1.15.2
k8s.gcr.io/kube-proxy:v1.15.2
k8s.gcr.io/pause:3.1
k8s.gcr.io/etcd:3.3.10
k8s.gcr.io/coredns:1.3.1
```

* 首先pull下来需要的images镜像：

```bash
>  kubeadm config images pull k8s.gcr.io/kube-apiserver:v1.15.2
>  kubeadm config images pull k8s.gcr.io/kube-controller-manager:v1.15.2
>  kubeadm config images pull k8s.gcr.io/kube-scheduler:v1.15.2
>  kubeadm config images pull k8s.gcr.io/kube-proxy:v1.15.2
>  kubeadm config images pull k8s.gcr.io/pause:3.1
>  kubeadm config images pull k8s.gcr.io/etcd:3.3.10
>  kubeadm config images pull k8s.gcr.io/coredns:1.3.1
```
也可以写个脚本安装:
```bash
#! /bin/bash

images=(
    kube-apiserver:v1.15.2
    kube-controller-manager:v1.15.2
    kube-scheduler:v1.15.2
    kube-proxy:v1.15.2
    pause:3.1
    etcd:3.3.15-0
    coredns:1.6.2
)
for imageName in ${images[@]} ; do
    docker pull registry.cn-hangzhou.aliyuncs.com/google_containers/${imageName}
    docker tag registry.cn-hangzhou.aliyuncs.com/google_containers/${imageName} k8s.gcr.io/${imageName}
    docker rmi registry.cn-hangzhou.aliyuncs.com/google_containers/${imageName}
done
```

* 执行初始化:

我们的Kubernetes主服务器已成功初始化！要开始使用群集，您需要以普通用户身份运行以下命令。

```bash
>  mkdir -p $HOME/.kube
> sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
> sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

现在，您应该将Pod网络部署到群集。使用以下列出的选项之一运行`kubectl apply -f [podnetwork] .yaml`：https://kubernetes.io/docs/concepts/cluster-administration/addons/

您可以通过运行以下命令来加入任意数量的计算机在每个节点上作为root：

```bash
> kubeadm join 192.168.79.130:6443 --token 1iq1mk.l366bc3w6nbv1vvd --discovery-token-ca-cert-hash sha256:c2c1d8ffb84733edd7991735ba9a20c4326adb146c58acd01287f8661a0d0bfb
```

如果遗忘：
```bash
> kubeadm token create --print-join-command
```

* 创建.kube目录

```bash
>  mkdir .kube 
```

* 复制admin.conf到.kube

```bash
> cp /etc/kubernetes/admin.conf .kube/config
```


然后需要部署[flannel网络插件](https://github.com/coreos/flannel).

* 拉取flannel的镜像

```bash
> docker pull quay.io/coreos/flannel:v0.11.0-amd64
```
* kubectl 安装kube-flannel
```bash
> kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
```

* 查看运行中pods

```bash
> kubectl get pods -n kube-system
NAME                             READY   STATUS    RESTARTS   AGE
coredns-86c58d9df4-k5xxp         1/1     Running   0          6m12s
coredns-86c58d9df4-k649n         1/1     Running   0          6m12s
etcd-master                      1/1     Running   0          5m29s
kube-apiserver-master            1/1     Running   0          5m29s
kube-controller-manager-master   1/1     Running   0          5m9s
kube-flannel-ds-amd64-jnk55      1/1     Running   0          47s
kube-proxy-xmdz5                 1/1     Running   0          6m12s
kube-scheduler-master            1/1     Running   0          5m26s
```

* 查看nodes

```bash
> kubectl get nodes
```


#### 13. node节点执行命令安装k8s

* node安装kubeadm

```bash
> yum install kubeadm kubelet -y
```
* 修改/etc/sysconfig/kubelet:

```bash
> vim /etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS="--fail-swap-on=false"
```

* 不用启动kubelet,直接加入集群


```bash
> kubeadm join 192.168.79.130:6443 --token 1iq1mk.l366bc3w6nbv1vvd --discovery-token-ca-cert-hash sha256:c2c1d8ffb84733edd7991735ba9a20c4326adb146c58acd01287f8661a0d0bfb
```

* 查看镜像:
```bash
> docker image list
REPOSITORY               TAG                 IMAGE ID            CREATED             SIZE
k8s.gcr.io/kube-proxy    v1.13.4             fadcc5d2b066        6 days ago          80.3MB
quay.io/coreos/flannel   v0.11.0-amd64       ff281650a721        5 weeks ago         52.6MB
k8s.gcr.io/pause         3.1                 da86e6ba6ca1        14 months ago       742kB
```
* 回到主节点查看节点信息过程中没有启动过kubelet

```bash
> kubectl get nodes
```

* 查看 kubectl配置信息

```bash
> kubectl config view
apiVersion: v1
clusters: []
contexts: []
current-context: ""
kind: Config
preferences: {}
users: []
```

* 配置从节点使用kubectl 查看节点信息
```bash
> mkdir .kube
> scp /etc/kubernetes/admin.conf node1:/root/.kube/config # 复制主节点下的配置到node
```
			
#### 14. 测试部署

* 测试nginx服务:

```bash
> kubectl run nginx --image=nginx --replicas=3
```

* 获取pod
```bash
> kubectl get pod
```

* 获取所有的pod
```bash
> kubectl get pod -o wide
```

* 部署服务
```bash
> kubectl expose deployment nginx --port=88 --target-port=80 --type=NodePort
```

* 查看服务信息
```bash
> kubectl get svc
```

其中的, `88:34710/TCP`也指定了一个`34710`端口，它表示可以通过node节点ip的这个端口访问服务。

这里有三个pod，请求不一定分配到哪一个pod上去了.

```bash
> kubectl logs nginx-8586cf59-bzmll
```
