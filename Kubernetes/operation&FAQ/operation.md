### k8s使用中遇到的问题和一些技巧

#### 1. 新加入节点，如何保证使用yum安装的是与之前版本一致
	1.1 当集群中新加入节点时，如果使用yum安装kubeadm，kubectl，kubelet并且不指定版本，模式是安装最新的版本，将会导致版本不一致的问题
		解决办法：  yum install -y kubelet-<version> kubectl-<version> kubeadm-<version>
		例如： yum install -y kubelet-1.16.1 kubeadm-1.16.1 kubectl-1.16.1

#### 2. kubeadm默认证书有效时间为1年，一年过期之后，会导致api server不可用
	错误提示： x509: certificate has expired or is not yet valid
	方案有如下：
	第一种：如果证书已经过期，解决办法
	2.1 如果证书已经过期： 针对kubeadm 1.13.x 及以上处理
	2.1.1 准备kubeadm.conf配置文件一份
		apiVersion: kubeadm.k8s.io/v1beta1
		kind: ClusterConfiguration
		kubernetesVersion: v1.14.1 #-->这里改成你集群对应的版本
		imageRepository: registry.cn-hangzhou.aliyuncs.com/google_containers 
		#这里使用国内的镜像仓库，否则在重新签发的时候会报错：could not fetch a Kubernetes version from the internet: unable to get URL "https://dl.k8s.io/release/stable-1.txt"
		文件位置：/root/kubeadm.conf
	2.1.2 重新签发命令
		kubeadm alpha certs renew all --config=/root/kubeadm.conf
		运行如上命令会重新生成以下证书 （临时解决办法，续约，1年）
		#-- /etc/kubernetes/pki/apiserver.key
		#-- /etc/kubernetes/pki/apiserver.crt
		
		#-- /etc/kubernetes/pki/apiserver-etcd-client.key
		#-- /etc/kubernetes/pki/apiserver-etcd-client.crt
		
		#-- /etc/kubernetes/pki/apiserver-kubelet-client.key
		#-- /etc/kubernetes/pki/apiserver-kubelet-client.crt
		
		#-- /etc/kubernetes/pki/front-proxy-client.key
		#-- /etc/kubernetes/pki/front-proxy-client.crt
		
		#-- /etc/kubernetes/pki/etcd/healthcheck-client.key
		#-- /etc/kubernetes/pki/etcd/healthcheck-client.crt
		
		#-- /etc/kubernetes/pki/etcd/peer.key
		#-- /etc/kubernetes/pki/etcd/peer.crt
		
		#-- /etc/kubernetes/pki/etcd/server.key
		#-- /etc/kubernetes/pki/etcd/server.crt

		完成后重启kube-apiserver,kube-controller,kube-scheduler,etcd这4个容器
	

	2.2 查看证书过期时间： kubeadm alpha certs check-expiration
	第二种： 开启自动轮换kubelet证书
	2.3 启用自动轮换kubelet 证书
		kubelet证书分为server和client两种， k8s 1.9默认启用了client证书的自动轮换，但server证书自动轮换需要用户开启
	2.3.1 增加 kubelet 参数
		# 在/etc/systemd/system/kubelet.service.d/10-kubeadm.conf 增加如下参数
		Environment="KUBELET_EXTRA_ARGS=--feature-gates=RotateKubeletServerCertificate=true"

	2.3.2 增加 controller-manager 参数
		
		# 在/etc/kubernetes/manifests/kube-controller-manager.yaml 添加如下参数
		  - command:
		    - kube-controller-manager
		    - --experimental-cluster-signing-duration=87600h0m0s
		    - --feature-gates=RotateKubeletServerCertificate=true
		    - ....
	2.3.3 创建 rbac 对象
		创建rbac对象，允许节点轮换kubelet server证书：
		cat > ca-update.yaml << EOF
		apiVersion: rbac.authorization.k8s.io/v1
		kind: ClusterRole
		metadata:
		  annotations:
		    rbac.authorization.kubernetes.io/autoupdate: "true"
		  labels:
		    kubernetes.io/bootstrapping: rbac-defaults
		  name: system:certificates.k8s.io:certificatesigningrequests:selfnodeserver
		rules:
		- apiGroups:
		  - certificates.k8s.io
		  resources:
		  - certificatesigningrequests/selfnodeserver
		  verbs:
		  - create
		---
		apiVersion: rbac.authorization.k8s.io/v1
		kind: ClusterRoleBinding
		metadata:
		  name: kubeadm:node-autoapprove-certificate-server
		roleRef:
		  apiGroup: rbac.authorization.k8s.io
		  kind: ClusterRole
		  name: system:certificates.k8s.io:certificatesigningrequests:selfnodeserver
		subjects:
		- apiGroup: rbac.authorization.k8s.io
		  kind: Group
		  name: system:nodes
		EOF
		
		kubectl create –f ca-update.yaml

	第三种： 修改源代码调整证书过期时间
		暂未测试，待测试完毕之后，补充上传

#### 3. 强制重启某个pod
	1. kubectl get pod coredns-5c98db65d4-jwqhv -n kube-system -o yaml | kubectl replace --force -f -

#### 4. kubeadm初始化之后抛出警告
	detected "cgroupfs" as the Docker cgroup driver. The recommended driver is "systemd"
	1. 解决办法：
	   	修改或创建/etc/docker/daemon.json，加入下面的内容：
		{
		  "exec-opts": ["native.cgroupdriver=systemd"]
		}
		
		systemctl reload-daemon
		systemctl restart docker

#### 5. 安装过程中coredns启动不起来，并且抛出日志
		"cni0" already has an IP address different from
		参考解决办法：
			kubeadm reset  由于集群此时还未搭建成功，建议kubeadm reset重置
			systemctl stop kubelet
			systemctl stop docker
			rm -rf /var/lib/cni/
			rm -rf /var/lib/kubelet/*
			rm -rf /etc/cni/
			ifconfig cni0 down
			ifconfig flannel.1 down
			ifconfig docker0 down
			ip link delete cni0
			ip link delete flannel.1
			systemctl start docker
		同理如果，新增节点在加入集群中出现意外，当你准备重新加入一直失败，可使用kubeadm reset 重置证书和一些配置文件

### 6. 在公司内网环境一般是不会将防火墙关闭，所以需要开启k8s使用到的端口
	centos7开放端口指定端口命令：
	4789/tcp、7946/tcp、7946/udp、2377/tcp、6443/tcp 10250-10252/tcp 2379-2380/tcp 10255/tcp 30000-32767/tcp
	firewall-cmd --zone=public --add-port 4789/tcp --permanent   TCP
	firewall-cmd --zone=public --add-port 7946/udp --permanent   UDP
	firewall-cmd  --reload

### 7. 配置Pod使用外部DNS
	修改kube-dns的使用的ConfigMap
	apiVersion: v1
	kind: ConfigMap
	metadata:
	  name: kube-dns
	  namespace: kube-system
	data:
	  stubDomains: |
	    {"k8s.com": ["192.168.196.18"]}
	  upstreamNameservers: |
	    ["8.8.8.8", "114.114.114.114"]

### 8. 强制删除一直处于Terminating状态的Pod
	出现该情况的原因可能有：与pod相关联的一些资源还未释放，再等待依赖的资源先释放
	1. 这种情况下，可以先删除与之相关联的资源，比如pvc,pv等
	2. 直接使用命令： kubectl delete pod $POD_ID --force --grace-period=0
			