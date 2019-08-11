在绝大多数生产场景下，k8s 集群对内，对外开启TLS属性。
这里为了演示一个完整的流程，使用自签发根证书对当前k8s集群进行通信加密和认证。本文档使用 CloudFlare 的 PKI 工具集 cfssl 创建所有证书。
为确保安全，kubernetes 系统各组件需要使用 x509 证书对通信进行加密和认证。

---
#### 安装cfssl 工具集

因为k8s 集群部署的便利，除强制要求以外，一般会将 cfssl 工具集安装到 k8s 具有master 属性的节点服务器上。因为此次演示均基于Linux系统环境下。所以使用 [cfssl linux 版本](http://pkg.cfssl.org/), 此演示环境中 master role 是主机名为 node01 的服务器。此部分为安装实施过程。
```yaml
# 切换至软件包下载安装临时文件夹下
[root@node1 ~]# cd /data/soft/
[root@node1 soft]# 
# 下载 cfssl 
[root@node1 soft]# wget https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
# 下载 cfssljson
[root@node1 soft]# wget https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
# 下载 cfssl-certinfo
[root@node1 soft]# wget https://pkg.cfssl.org/R1.2/cfssl-certinfo_linux-amd64
# 当前下载文件列表清单
[root@node1 soft]# ll
总用量 18808
-rw-r--r-- 1 root root  6595195 3月  30 2016 cfssl-certinfo_linux-amd64
-rw-r--r-- 1 root root  2277873 3月  30 2016 cfssljson_linux-amd64
-rw-r--r-- 1 root root 10376657 3月  30 2016 cfssl_linux-amd64
# 将对应的文件重命名并分配给k8s 用户及授予可执行权限
[root@node1 soft]# mv cfssl-certinfo_linux-amd64 cfssl-certinfo 
[root@node1 soft]# mv cfssljson_linux-amd64 cfssljson
[root@node1 soft]# mv cfssl_linux-amd64 cfssl
[root@node1 soft]# 
[root@node1 soft]# chmod +x cfssl*
[root@node1 soft]# chown -R k8s cfssl*
[root@node1 soft]# ll
总用量 18808
-rwxr-xr-x 1 k8s root 10376657 3月  30 2016 cfssl
-rwxr-xr-x 1 k8s root  6595195 3月  30 2016 cfssl-certinfo
-rwxr-xr-x 1 k8s root  2277873 3月  30 2016 cfssljson
# 将该执行脚本拷贝到 k8s 的可执行文件(bin)路径下
[root@node1 soft]# cp -p cfssl* /usr/local/k8s/bin/
[root@node1 soft]# cd /usr/local/k8s/bin/
[root@node1 bin]# ll
总用量 18808
-rwxr-xr-x 1 k8s root 10376657 3月  30 2016 cfssl
-rwxr-xr-x 1 k8s root  6595195 3月  30 2016 cfssl-certinfo
-rwxr-xr-x 1 k8s root  2277873 3月  30 2016 cfssljson
```

效果校验

```yaml
# 查看 cfssl 当前版本
[root@node1 ~]# cfssl version
Version: 1.2.0
Revision: dev
Runtime: go1.6
# cfssljson flag
[root@node1 ~]# cfssljson -h
Usage of cfssljson:
  -bare
        the response from CFSSL is not wrapped in the API standard response
  -f string
        JSON input (default "-")
  -stdout
        output the response instead of saving to a file
```

---

#### 创建 ***CA***

[数字证书认证机构](https://zh.wikipedia.org/wiki/%E8%AF%81%E4%B9%A6%E9%A2%81%E5%8F%91%E6%9C%BA%E6%9E%84)（英语：Certificate Authority，缩写为CA），也称为电子商务认证中心、电子商务认证授权机构，是负责发放和管理数字证书的权威机构，并作为电子商务交易中受信任的第三方，承担公钥体系中公钥的合法性检验的责任。
CA 证书的配置文件，用于配置根证书的使用场景 (profile) 和具体参数 (usage，过期时间、服务端认证、客户端认证、加密等)，后续在签名其它证书时需要指定特定场景。 这里为了演示，集群所有节点共享的，***只需要创建一个 CA 证书***，后续创建的所有证书都由它签名。
```json
cat > ca-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "87600h"
    },
    "profiles": {
      "kubernetes": {
        "usages": [
            "signing",
            "key encipherment",
            "server auth",
            "client auth"
        ],
        "expiry": "87600h"
      }
    }
  }
}
EOF
```
### ***备注说明:***
```shell
signing：表示该证书可用于签名其它证书，生成的 ca.pem 证书中 CA=TRUE；  
server auth：表示 client 可以用该该证书对 server 提供的证书进行验证；  
client auth：表示 server 可以用该该证书对 client 提供的证书进行验证；
```
---

#### 创建证书签名请求文件

````json
cat > ca-csr.json <<EOF
{
  "CN": "kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "ST": "BeiJing",
      "L": "BeiJing",
      "O": "k8s",
      "OU": "Demo"
    }
  ]
}
EOF
````
检查内容及生成对应的证书签名
```shell
[root@node1 soft]# pwd
/data/soft
[root@node1 soft]# ll *.json
-rw-r--r-- 1 root root 292 8月  11 01:28 ca-config.json
-rw-r--r-- 1 root root 206 8月  11 01:29 ca-csr.json
[root@node1 soft]# cfssl gencert -initca ca-csr.json | cfssljson -bare ca
2019/08/11 01:32:13 [INFO] generating a new CA key and certificate from CSR
2019/08/11 01:32:13 [INFO] generate received request
2019/08/11 01:32:13 [INFO] received CSR
2019/08/11 01:32:13 [INFO] generating key: rsa-2048
2019/08/11 01:32:14 [INFO] encoded CSR
2019/08/11 01:32:14 [INFO] signed certificate with serial number 626829679744097151039756577561932032493741610295
# 检查生成的内容及文件
[root@node1 soft]# ls ca*
ca-config.json  ca.csr  ca-csr.json  ca-key.pem  ca.pem
```

### ***参数说明:***
```
CN：Common Name，kube-apiserver 从证书中提取该字段作为请求的用户名 (User Name)，浏览器使用该字段验证网站是否合法；
O：Organization，kube-apiserver 从证书中提取该字段作为请求用户所属的组 (Group)；
kube-apiserver 将提取的 User、Group 作为 RBAC 授权的用户标识；
```
---

#### 分发证书文件
将生成的 CA 证书、秘钥文件、配置文件拷贝到所有master节点和node节点的 /etc/kubernetes/cert 目录下,并保证k8s用户有读写 /etc/kubernetes 目录及其子目录文件的权限：
```shell
[root@node1 soft]# cp ca*.pem ca-config.json /etc/kubernetes/cert/
[root@node1 soft]# chown -R k8s /etc/kubernetes/cert/ca*
[root@node1 soft]# cd /etc/kubernetes/cert/
[root@node1 cert]# ll
总用量 12
-rw-r--r-- 1 k8s root  292 8月  11 01:37 ca-config.json
-rw------- 1 k8s root 1675 8月  11 01:37 ca-key.pem
-rw-r--r-- 1 k8s root 1354 8月  11 01:37 ca.pem
```