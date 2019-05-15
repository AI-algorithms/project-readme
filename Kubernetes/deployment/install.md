# kubernetes 部署文档

### 集群架构
![kubernetes-Architecture](/Kubernetes/images/kubernetes-Architecture.png)

## 组件版本
序 号 |  组件名称  | 版本号  | 备 注
:---: | :--------: | :------: | :----:
1   |  Kubernetes  | 1.12.0-rc2 | 
2   |  Docker   | 18.06.1-ce    | 
3   | Etcd  | 3.3.7 | 
4   | Flanneld  | 0.11.0 | 

+ 插件：
    - Coredns
    - Dashboard
    - Heapster (influxdb、grafana)
    - Metrics-Server
    - ERK (elasticsearch、Rsyslog、kibana)
+ 镜像仓库：
    - docker registry
    - harbor


## 主要配置策略

***kube-apiserver：***

+ 使用 keepalived 和 haproxy 实现 3 节点高可用；
+ 关闭非安全端口 8080 和匿名访问；
+ 在安全端口 6443 接收 https 请求；
+ 严格的认证和授权策略 (x509、token、RBAC)；
+ 开启 bootstrap token 认证，支持 kubelet TLS bootstrapping；
+ 使用 https 访问 kubelet、etcd，加密通信；

***kube-controller-manager：***

+ 3 节点高可用；
+ 关闭非安全端口，在安全端口 10252 接收 https 请求；
+ 使用 kubeconfig 访问 apiserver 的安全端口；
+ 自动 approve kubelet 证书签名请求 (CSR)，证书过期后自动轮转；
+ 各 controller 使用自己的 ServiceAccount 访问 apiserver；

kube-scheduler：

+ 3 节点高可用；
+ 使用 kubeconfig 访问 apiserver 的安全端口；

***kubelet：***

+ 使用 kubeadm 动态创建 bootstrap token，而不是在 apiserver 中静态配置；
+ 使用 TLS bootstrap 机制自动生成 client 和 server 证书，过期后自动轮转；
+ 在 KubeletConfiguration 类型的 JSON 文件配置主要参数；
+ 关闭只读端口，在安全端口 10250 接收 https 请求，对请求进行认证和授权，拒绝匿名访问和非授权访问；
+ 使用 kubeconfig 访问 apiserver 的安全端口；

***kube-proxy：***

+ 使用 kubeconfig 访问 apiserver 的安全端口；
+ 在 KubeProxyConfiguration  类型的 JSON 文件配置主要参数；
+ 使用 ipvs 代理模式；

*** etcd: ***
+ 3 节点高可用
+ etcd 启用 https