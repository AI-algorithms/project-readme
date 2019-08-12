### 使用Kubeadm搭建kubernetes集群

#### 前期准备
  
  - [Docker安装, 选择适配操作系统的版本](<https://docs.docker.com/install/linux/docker-ce/centos/>)
  
  - [Kubectl, Kubelet, Kubeadm安装](<https://kubernetes.io/zh/docs/setup/independent/install-kubeadm/>)
  
  - Kubectl, Kubelet, Kubeadm墙内安装, 我这里是centos, 不同操作系统查找相应的安装配置方式
  
    ```bash
    cat <<EOF > /etc/yum.repos.d/kubernetes.repo
    [kubernetes]
    name=Kubernetes
    baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
    enabled=1
    gpgcheck=1
    repo_gpgcheck=1
    gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
    EOF
    # 安装
    yum install -y kubectl kubelet kubeadm
    # 开机启动
    systemctl enable kubelet
    # 启动
    systemctl start kubelet
    ```
  
  - [配置docker,Cgroup驱动, 以及阿里云镜像加速地址](<https://kubernetes.io/docs/setup/production-environment/container-runtimes/>)
  
    ```bash
    # Setup daemon.
    cat > /etc/docker/daemon.json <<EOF
    {
    "registry-mirrors": ["https://xxxx.mirror.aliyuncs.com"],
      "exec-opts": ["native.cgroupdriver=systemd"],
      "log-driver": "json-file",
      "log-opts": {
        "max-size": "100m"
      },
      "storage-driver": "overlay2",
      "storage-opts": [
        "overlay2.override_kernel_check=true"
      ]
    }
    ```
  
  - 拉取所需镜像
  
    ```
    // 拉取阿里云镜像仓库中的公共镜像
    sudo docker pull registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-controller-manager:v1.14.1
    sudo docker pull registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-apiserver:v1.14.1
    sudo docker pull registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-scheduler:v1.14.1
    sudo docker pull registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-proxy:v1.14.1
    sudo docker pull registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/pause:3.1
    sudo docker pull registry.cn-hangzhou.aliyuncs.com/jxqc/etcd:3.3.10
    sudo docker pull registry.cn-hangzhou.aliyuncs.com/jxqc/coredns:1.3.1
    sudo docker pull registry.cn-hangzhou.aliyuncs.com/kuberneters/kubernetes-dashboard-amd64:v1.10.1
    // 为拉取的镜像重新打tag
    sudo docker tag registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-controller-manager:v1.14.1     k8s.gcr.io/kube-controller-manager:v1.14.1
    sudo docker tag registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-apiserver:v1.14.1                          k8s.gcr.io/kube-apiserver:v1.14.1
    sudo docker tag registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-scheduler:v1.14.1                         k8s.gcr.io/kube-scheduler:v1.14.1
    sudo docker tag registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-proxy:v1.14.1                                  k8s.gcr.io/kube-proxy:v1.14.1
    sudo docker tag registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/pause:3.1                                                      k8s.gcr.io/pause:3.1
    sudo docker tag registry.cn-hangzhou.aliyuncs.com/jxqc/etcd:3.3.10                                                                                                 k8s.gcr.io/etcd:3.3.10
    sudo docker tag registry.cn-hangzhou.aliyuncs.com/jxqc/coredns:1.3.1                                                                                            k8s.gcr.io/coredns:1.3.1
    sudo docker tag registry.cn-hangzhou.aliyuncs.com/kuberneters/kubernetes-dashboard-amd64:v1.10.1                      k8s.gcr.io/kubernetes-dashboard-amd64:v1.10.1
    
    // 删除无用的镜像
    sudo docker rmi registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-controller-manager:v1.14.1
    sudo docker rmi registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-apiserver:v1.14.1
    sudo docker rmi registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-scheduler:v1.14.1
    sudo docker rmi registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/kube-proxy:v1.14.1
    sudo docker rmi registry.cn-hangzhou.aliyuncs.com/mirror_google_containers/pause:3.1
    sudo docker rmi registry.cn-hangzhou.aliyuncs.com/jxqc/etcd:3.3.10
    sudo docker rmi registry.cn-hangzhou.aliyuncs.com/jxqc/coredns:1.3.1
    sudo docker rmi registry.cn-hangzhou.aliyuncs.com/kuberneters/kubernetes-dashboard-amd64:v1.10.1
    ```
  
  - **以上相同配置的子节点一台**
  
  - 配置Master节点的Kubelet驱动, 修改 --cgroup-driver=cgroupfs  为  --cgroup-driver=systemd
  
    `sudo vi /etc/systemd/system/kubelet.service.d/10-kubeadm.conf`
  
    ```bash
    [Unit]
    Wants=docker.socket
    
    [Service]
    ExecStart=
    ExecStart=/usr/bin/kubelet --allow-privileged=true --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cgroup-driver=systemd --client-ca-file=/etc/kubernetes/pki/ca.crt --cluster-dns=10.96.0.10 --cluster-domain=cluster.local --container-runtime=docker --fail-swap-on=false --kubeconfig=/etc/kubernetes/kubelet.conf --pod-manifest-path=/etc/kubernetes/manifests
    
    [Install]
    ```
  
  - Master节点初始化
  
    `sudo kubeadm init --apiserver-advertise-address 192.168.0.109 --pod-network-cidr=10.244.0.0/16`
  
    ```
    I0619 21:09:34.578805   30353 version.go:96] could not fetch a Kubernetes version from the internet: unable to get URL "https://dl.k8s.io/release/stable-1.txt": Get https://dl.k8s.io/release/stable-1.txt: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)
    I0619 21:09:34.578939   30353 version.go:97] falling back to the local client version: v1.14.1
    [init] Using Kubernetes version: v1.14.1
    [preflight] Running pre-flight checks
    [preflight] Pulling images required for setting up a Kubernetes cluster
    [preflight] This might take a minute or two, depending on the speed of your internet connection
    [preflight] You can also perform this action in beforehand using 'kubeadm config images pull'
    [kubelet-start] Writing kubelet environment file with flags to file "/var/lib/kubelet/kubeadm-flags.env"
    [kubelet-start] Writing kubelet configuration to file "/var/lib/kubelet/config.yaml"
    [kubelet-start] Activating the kubelet service
    [certs] Using certificateDir folder "/etc/kubernetes/pki"
    [certs] Generating "front-proxy-ca" certificate and key
    [certs] Generating "front-proxy-client" certificate and key
    [certs] Generating "etcd/ca" certificate and key
    [certs] Generating "etcd/peer" certificate and key
    [certs] etcd/peer serving cert is signed for DNS names [chen localhost] and IPs [192.168.0.109 127.0.0.1 ::1]
    [certs] Generating "apiserver-etcd-client" certificate and key
    [certs] Generating "etcd/server" certificate and key
    [certs] etcd/server serving cert is signed for DNS names [chen localhost] and IPs [192.168.0.109 127.0.0.1 ::1]
    [certs] Generating "etcd/healthcheck-client" certificate and key
    [certs] Generating "ca" certificate and key
    [certs] Generating "apiserver" certificate and key
    [certs] apiserver serving cert is signed for DNS names [chen kubernetes kubernetes.default kubernetes.default.svc kubernetes.default.svc.cluster.local] and IPs [10.96.0.1 192.168.0.109]
    [certs] Generating "apiserver-kubelet-client" certificate and key
    [certs] Generating "sa" key and public key
    [kubeconfig] Using kubeconfig folder "/etc/kubernetes"
    [kubeconfig] Writing "admin.conf" kubeconfig file
    [kubeconfig] Writing "kubelet.conf" kubeconfig file
    [kubeconfig] Writing "controller-manager.conf" kubeconfig file
    [kubeconfig] Writing "scheduler.conf" kubeconfig file
    [control-plane] Using manifest folder "/etc/kubernetes/manifests"
    [control-plane] Creating static Pod manifest for "kube-apiserver"
    [control-plane] Creating static Pod manifest for "kube-controller-manager"
    [control-plane] Creating static Pod manifest for "kube-scheduler"
    [etcd] Creating static Pod manifest for local etcd in "/etc/kubernetes/manifests"
    [wait-control-plane] Waiting for the kubelet to boot up the control plane as static Pods from directory "/etc/kubernetes/manifests". This can take up to 4m0s
    [apiclient] All control plane components are healthy after 16.003603 seconds
    [upload-config] storing the configuration used in ConfigMap "kubeadm-config" in the "kube-system" Namespace
    [kubelet] Creating a ConfigMap "kubelet-config-1.14" in namespace kube-system with the configuration for the kubelets in the cluster
    [upload-certs] Skipping phase. Please see --experimental-upload-certs
    [mark-control-plane] Marking the node chen as control-plane by adding the label "node-role.kubernetes.io/master=''"
    [mark-control-plane] Marking the node chen as control-plane by adding the taints [node-role.kubernetes.io/master:NoSchedule]
    [bootstrap-token] Using token: qg3ci5.kvvoglj9lxejc3ij
    [bootstrap-token] Configuring bootstrap tokens, cluster-info ConfigMap, RBAC Roles
    [bootstrap-token] configured RBAC rules to allow Node Bootstrap tokens to post CSRs in order for nodes to get long term certificate credentials
    [bootstrap-token] configured RBAC rules to allow the csrapprover controller automatically approve CSRs from a Node Bootstrap Token
    [bootstrap-token] configured RBAC rules to allow certificate rotation for all node client certificates in the cluster
    [bootstrap-token] creating the "cluster-info" ConfigMap in the "kube-public" namespace
    [addons] Applied essential addon: CoreDNS
    [addons] Applied essential addon: kube-proxy
    
    Your Kubernetes control-plane has initialized successfully!
    
    To start using your cluster, you need to run the following as a regular user:
    
      mkdir -p $HOME/.kube
      sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
      sudo chown $(id -u):$(id -g) $HOME/.kube/config
    
    You should now deploy a pod network to the cluster.
    Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
      https://kubernetes.io/docs/concepts/cluster-administration/addons/
    
    Then you can join any number of worker nodes by running the following on each as root:
    
    kubeadm join 192.168.0.109:6443 --token qg3ci5.kvvoglj9lxejc3ij \
        --discovery-token-ca-cert-hash sha256:3a2af178182529a7e163260ec80f546edce98213c252cb0e6fdc0158a5cc876a 
    ```
  
  - Master切换到CKL
  
    ```bash
    mv  $HOME/.kube $HOME/.kube.bak
    mkdir -p $HOME/.kube
    sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
    sudo chown $(id -u):$(id -g) $HOME/.kube/config
    ```
  
  - 开启flannel网络插件
  
    `kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml`
  
  - 开启dashboard服务
  
    `kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v1.10.1/src/deploy/recommended/kubernetes-dashboard.yaml`
  
  - 子节点加入主节点集群
  
    `kubeadm join 192.168.0.109:6443 --token qg3ci5.kvvoglj9lxejc3ij    --discovery-token-ca-cert-hash sha256:3a2af178182529a7e163260ec80f546edce98213c252cb0e6fdc0158a5cc876a `
  
  - 主节点查看集群
  
    ```bash
    chen@chen:~$ sudo kubectl get nodes
    NAME    STATUS   ROLES    AGE   VERSION
    bogon   Ready    <none>   45m   v1.14.3
    chen    Ready    master   71m   v1.14.1
    ```
  
  - 查看集群中的pod,本地nameserver 127.0.0.53导致coredns的CrashLoopBackOff
  
    ```bash
    chen@chen:~/kubernetes$ sudo kubectl get pod --all-namespaces
    NAMESPACE     NAME                                    READY   STATUS             RESTARTS   AGE
    kube-system   coredns-fb8b8dccf-rjjfh                 0/1     CrashLoopBackOff   9          21m
    kube-system   coredns-fb8b8dccf-vgrx5                 0/1     CrashLoopBackOff   9          21m
    kube-system   etcd-chen                               1/1     Running            0          20m
    kube-system   kube-apiserver-chen                     1/1     Running            0          20m
    kube-system   kube-controller-manager-chen            1/1     Running            0          20m
    kube-system   kube-flannel-ds-amd64-8hf89             1/1     Running            0          20m
    kube-system   kube-flannel-ds-amd64-vs5pv             1/1     Running            0          17m
    kube-system   kube-proxy-mrvgk                        1/1     Running            0          17m
    kube-system   kube-proxy-pbwjw                        1/1     Running            0          21m
    kube-system   kube-scheduler-chen                     1/1     Running            0          20m
    kube-system   kubernetes-dashboard-5f7b999d65-jgnrq   1/1     Running            0          11m
    ```
  
  - [创建一个amdin-user, 并绑定一个管理员角色](<https://github.com/kubernetes/dashboard/wiki/Creating-sample-user>)
  
    -  `vi  dashboard-adminuser-create.yaml`
  
    ```yaml
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: admin-user
      namespace: kube-system
    ```
  
    `创建用户:sudo kubectl apply -f dashboard-adminuser-create.yaml`
  
    - `vi dashboard-adminuser-role-binding.yaml`
  
    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: admin-user
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: cluster-admin
    subjects:
    - kind: ServiceAccount
      name: admin-user
      namespace: kube-system
    ```
  
    `为用户绑定角色: sudo kubectl apply -f dashboard-adminuser-role-binding.yaml`
  
  - 启动dashboard
  
    ```
    sudo kubectl proxy                                                                                       --  只能本地访问
    sudo kubectl proxy --address='0.0.0.0' --accept-hosts='^*$'     --  外部可访问
    ```
  
  - 生成admin-user的登录令牌
  
    `sudo kubectl -n kube-system describe secret $(kubectl -n kube-system get secret | grep admin-user | awk '{print $1}')`
  
  - 访问dashboard
  
    `http://localhost:8001/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy/#!/login`
