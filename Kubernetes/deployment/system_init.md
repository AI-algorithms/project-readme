
### 部署角色分配列表

| Hostname  | IP | Roles |
| :------------- | :------------- | :------------- |
| node01  | 192.168.50.2  | kube-apiserver kube-controllermanager kube-scheduler etcd |
| node02  | 192.168.50.3  | kube-apiserver kube-controllermanager kube-scheduler etcd |
| node03  | 192.168.50.4  | kube-apiserver kube-controllermanager kube-scheduler etcd |
| node04  | 192.168.50.181  | kubelet |
| node06  | 192.168.50.171  | kubelet |

- 基础设施环境

#### 系统及内核版本
```bash
# 操作系统版本
[root@node01 ~]# cat /etc/redhat-release 
CentOS Linux release 7.5.1804 (Core)
# 系统内核版本
[root@node01 ~]# uname -mrs
Linux 5.1.10-1.el7.elrepo.x86_64 x86_64
```
------

#### 为保证环境统一，参考如 **部署角色分配列表** 所示，修改对应服务器的主机名称及追加主机名与IP地址的映射关系，安装集群正常运行所需的系统依赖，关闭服务器本地的防火墙服务及日常模块安装优化。
-  修改服务器的主机名
```bash
# node01
[root@node1 ~]# hostnamectl --static set-hostname node1
# node02
[root@node2 ~]# hostnamectl --static set-hostname node2
# node03
[root@node3 ~]# hostnamectl --static set-hostname node3
# node04
[root@node4 ~]# hostnamectl --static set-hostname node4
# node06
[root@node6 ~]# hostnamectl --static set-hostname node6
```
- 追加集群间主机名解析

> 这是为了演示，统一使用服务器名称为 **node01** 作为演示
```bash
[root@node01 ~]# cat  >> /etc/hosts <<EOF
192.168.50.2    node01
192.168.50.3    node02
192.168.50.4    node03
192.168.50.181  node04
192.168.50.171  node06
EOF
```

- 安装系统依赖
```bash
[root@node01 ~]# yum install -y epel-release
[root@node01 ~]# yum install -y conntrack ipvsadm ipset jq sysstat curl iptables libseccomp
```

- 关闭防火墙  
```bash
# 关闭防火墙
[root@node01 ~]# systemctl stop firewalld
# 禁用服务器开机自启动服务
[root@node01 ~]# systemctl disable firewalld
# 清空 iptables 规则栈
[root@node01 ~]# iptables -F &&  iptables -X &&  iptables -F -t nat &&  iptables -X -t nat
[root@node01 ~]# iptables -P FORWARD ACCEPT

# 验证iptables 是否清空
[root@node01 ~]# iptables -L -n 
Chain INPUT (policy ACCEPT)
target     prot opt source               destination         

Chain FORWARD (policy ACCEPT)
target     prot opt source               destination         

Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination 
```
- **验证防火墙是否关闭**
```bash
[root@node01 ~]# firewall-cmd --state
not running
```

- 关闭 swap 分区
```bash
# 不禁用swap的话，在启动kubelet 或者 docker 都会有问题
[root@node01 ~]# swapoff -a
# 验证swap 是否关闭
[root@node01 ~]# cat /etc/fstab  | grep -w swap
#/dev/mapper/centos-swap swap                    swap    defaults        0 0
```
> 为了防止开机自动挂载 swap 分区，可以注释 /etc/fstab 中相应的条目：
````bash
[root@node01 ~]# sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab
[root@node01 ~]#
````

-  关闭 **SELINUX**
```bash
# 必须要关闭 selinux, 否则后续 K8S 挂载目录时可能报错 Permission denied：
[root@node01 ~]# setenforce 0
setenforce: SELinux is disabled
# 或者也可以执行以下命名行关闭
# sed -i --follow-symlinks 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/sysconfig/selinux
# 验证是否关闭
[root@node01 ~]# grep -w SELINUX /etc/selinux/config
# SELINUX= can take one of these three values:
SELINUX=disabled
```

- 安装网络模块
```bash
# 这里使用 flannel 作为k8s 的网络通信插件
[root@node01 ~]# modprobe bridge
# 如果内核版本低于 3.10.0-327.36.2.el7.x86_64, 是无法加载 br_netfilter 网络模块
[root@node01 ~]# modprobe br_netfilter
[root@node01 ~]# echo '1' > /proc/sys/net/bridge/bridge-nf-call-iptables
[root@node01 ~]# modprobe ip_vs
```

- 设置k8s 服务的系统参数
```bash
[root@node1 ~]# cat >>  /etc/sysctl.d/kubernetes.conf <<EOF
net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
net.ipv4.ip_forward=1
net.ipv4.tcp_tw_recycle=0
vm.swappiness=0
vm.overcommit_memory=1
vm.panic_on_oom=0
fs.inotify.max_user_watches=89100
fs.file-max=52706963
fs.nr_open=52706963
net.ipv6.conf.all.disable_ipv6=1
net.netfilter.nf_conntrack_max=2310720
EOF
# 加载服务所需的内核参数
[root@node1 sysctl.d]# sysctl -p /etc/sysctl.d/kubernetes.conf 
net.bridge.bridge-nf-call-iptables = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward = 1
sysctl: cannot stat /proc/sys/net/ipv4/tcp_tw_recycle: No such file or directory
vm.swappiness = 0
vm.overcommit_memory = 1
vm.panic_on_oom = 0
fs.inotify.max_user_watches = 89100
fs.file-max = 52706963
fs.nr_open = 52706963
net.ipv6.conf.all.disable_ipv6 = 1
net.netfilter.nf_conntrack_max = 2310720


[root@node1 ~]# mount -t cgroup -o cpu,cpuacct none /sys/fs/cgroup/cpu,cpuacct
```
------
-  设置系统时区
```bash
# 调整服务器的时间 TimeZone
[root@node1 ~]# timedatectl set-timezone Asia/Shanghai
# 更新服务器的时间
[root@node1 ~]# ntpdate cn.pool.ntp.org
15 Jun 23:23:08 ntpdate[13173]: step time server 119.28.206.193 offset 371.235820 sec
# 当前的 UTC 时间写入硬件时钟
[root@node1 ~]# timedatectl set-local-rtc 0
# 重启依赖于服务器系统时间的服务
[root@node1 ~]# systemctl restart rsyslog
[root@node1 ~]# systemctl restart crond
```
-----
- 添加 **k8s** 账户
```bash
[root@node1 ~]# useradd -m k8s
# 这里只是为了演示，将k8s用户的密码设置为 123456
[root@node1 ~]# sh -c 'echo 123456 | passwd k8s --stdin'
Changing password for user k8s.
passwd: all authentication tokens updated successfully.
# 为了提高安全性
[root@node1 ~]# chsh k8s -s /sbin/nologin
Changing shell for k8s.
Shell changed.
```

-----

-  创建工作目录
```bash
# 创建k8s 服务服务执行空间
[root@node1 ~]# mkdir -p /usr/local/k8s/bin
[root@node1 ~]# chown -R k8s /usr/local/k8s

# 创建k8s 服务存放认证文件空间
[root@node1 ~]# mkdir -p /etc/kubernetes/cert
[root@node1 ~]# chown -R k8s /etc/kubernetes

# 因为在该演示环境中， etcd 集群是跟 k8s master 服务混用，所以以下操作是在 master 节点上操作。 work data 不需要操作执行。
[root@node1 ~]# mkdir -p /etc/etcd/cert
[root@node1 ~]# mkdir -p /var/lib/etcd && chown -R k8s /etc/etcd/cert

```
----
> 在下一步启动组件时，很可能会因为权限不对，导致服务启动失败，要注意给相关目录设置属主为 k8s
```bash
# 这里拿 kubernetes 服务的文件夹目录作为演示
[root@node1 ~]# cd /etc/kubernetes/
[root@node1 kubernetes]# ll
total 0
drwxr-xr-x 2 k8s root 6 Jun 15 23:32 cert
```
-----
- 非必须

***将可执行文件路径 /usr/local/k8s/bin 添加到 系统 PATH 变量中***
```
[root@node1 ~]# sh -c "echo 'export PATH=/usr/local/k8s/bin:$PATH' >> ~/.bashrc"
[root@node1 ~]# source ~/.bashrc 
```
-----

#### 内核升级
默认情况下，centos 7 内核版本为 3.10.0-327.el7.x86_64 x86_64，在这个版本下加载网络模块是无法加载的。需要升级系统内核来解决。

***内核升级有风险，这里只做展示***

***内核操作有风险，请谨慎操作***
```bash
[root@node1 ~]# wget https://www.elrepo.org/RPM-GPG-KEY-elrepo.org
[root@node1 ~]# rpm --import RPM-GPG-KEY-elrepo.org
[root@node1 ~]# wget http://www.elrepo.org/elrepo-release-7.0-2.el7.elrepo.noarch.rpm
[root@node1 ~]# rpm -ihv elrepo-release-7.0-2.el7.elrepo.noarch.rpm
Retrieving http://elrepo.org/elrepo-release-7.0-3.el7.elrepo.noarch.rpm
Preparing...                          ################################# [100%]
Updating / installing...
   1:elrepo-release-7.0-3.el7.elrepo  ################################# [100%]

# 查看可升级内核模块及相应版本
[root@node1 ~]# yum list available --disablerepo='*' --enablerepo=elrepo-kernel
# 升级内核版本，型号为kernel-lt 
[root@node1 ~]# yum --disablerepo='*' --enablerepo=elrepo-kernel install kernel-lt
# 查看服务器当前已安装的内核版本
[root@node1 ~]# awk -F\' '$1=="menuentry " {print i++ " : " $2}' /etc/grub2.cfg
0 : CentOS Linux (5.1.10-1.el7.elrepo.x86_64) 7 (Core)
1 : CentOS Linux (4.4.181-1.el7.elrepo.x86_64) 7 (Core)
2 : CentOS Linux (3.10.0-327.36.2.el7.x86_64) 7 (Core)
3 : CentOS Linux (3.10.0-327.el7.x86_64) 7 (Core)
4 : CentOS Linux (0-rescue-e20630b279f74ade9b23049f62de9be4) 7 (Core)
# 选择你所需要的升级后的内核版本序号
# 这里选择内核版本为 4.4.181-1.el7.elrepo.x86_64，当前序号为 1
[root@node1 ~]# grub2-set-default 1
[root@node1 ~]# grub2-mkconfig -o /boot/grub2/grub.cfg
Generating grub configuration file ...
Found linux image: /boot/vmlinuz-4.4.181-1.el7.elrepo.x86_64
Found initrd image: /boot/initramfs-4.4.181-1.el7.elrepo.x86_64.img
Found linux image: /boot/vmlinuz-3.10.0-327.el7.x86_64
Found initrd image: /boot/initramfs-3.10.0-327.el7.x86_64.img
Found linux image: /boot/vmlinuz-0-rescue-9d8b479c072b4b709f271c3126e1322f
Found initrd image: /boot/initramfs-0-rescue-9d8b479c072b4b709f271c3126e1322f.img
[root@node1 ~]# reboot
```