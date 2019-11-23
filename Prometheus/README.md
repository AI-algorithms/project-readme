#### prometheus结合grafana的使用

业务需求:

* 多维度满足业务项目的需求
* 统一监控系统
* 满足云原生需求

集群规划:

![design.architecture](/Prometheus/images/ops.prometheus.design.architecture.png)

组件版本:

| name | version|  path | 
| :---: | :---: | :---: |
| `prometheus` | 2.13.1| /data/app/prometheus-2.13.1 |
| `alertmanager` | 0.18.0 | /data/app/alertmanager-0.18.0 |
| `grafana` | 6.3.3 |  rpm |
| `consul` | 1.6.2 | /usr/local/bin | 


#### 部署安装prometheus

```bash
cd /data/soft
wget https://github.com/prometheus/prometheus/releases/download/v2.13.1/prometheus-2.13.1.linux-amd64.tar.gz

tar xzf prometheus-2.13.1.linux-amd64.tar.gz
mv prometheus-2.13.1.linux-amd64 prometheus-2.13.1
mv prometheus-2.13.1 ../app/
```

* 实例systemctl服务(instance systemctl service)

```bash
cat > /usr/lib/systemd/system/prometheus.service <<EOF

[Unit]
Description=Prometheus Server
Documentation=https://prometheus.io/docs/introduction/overview/
After=network-online.target

[Service]
LimitNOFILE=1048576
LimitNPROC=1048576
LimitCORE=infinity
User=root
Restart=on-failure
ExecStart=/data/app/prometheus-2.13.1/bin/prometheus \\
                --config.file=/data/app/prometheus-2.13.1/conf.d/prometheus.yml \\
                --storage.tsdb.path=/data/app/prometheus-2.13.1/data \\
                --web.listen-address=:9090 --web.enable-lifecycle  \\
                --storage.remote.flush-deadline=1h \\
                --web.console.templates=/data/app/prometheus-2.13.1/consoles  \\
                --web.console.libraries=/data/app/prometheus-2.13.1/console_libraries

ExecReload=/bin/kill -HUP $MAINPID
[Install]
WantedBy=multi-user.target
EOF
```

* 实例配置(instance configuration)

children instance configuartion:

```yaml
global:
scrape_interval:     5s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.
scrape_timeout: 5s # is set to the global default (10s).
external_labels:
  dc: 'dc01'

alerting:
alertmanagers:
- static_configs:
    - targets: [ 'alertmanagerList' ]

scrape_configs:
- job_name: 'dc01-children-prometheus'
    static_configs:
    - targets: [ 'dc01-children-prometheus:9090' ]

- job_name: 'serviceauto'
    scrape_interval: 30s
    scrape_timeout: 20s
    static_configs:
    consul_sd_configs:
    - server: 'consulClientName:Port' # not suports Lists
        scheme: 'http'

    relabel_configs:
    - source_labels: ['__meta_consul_service']
        regex:         '(.*)'
        target_label:  'job'
        replacement:   '$1'
    - source_labels: ['__address__']
        regex:         '(.*)'
        target_label:  'service'
        replacement:   '$1'
    - source_labels: ['__meta_consul_service_metadata_project']
        regex:         '(.*)'
        target_label:  'project'
        replacement:   '$1'
    - source_labels: ['__meta_consul_service_metadata_role']
        regex:         '(.*)'
        target_label:  'role'
        replacement:   '$1'
    - source_labels: ['__meta_consul_service_metadata_envior']
        regex:         '(.*)'
        target_label:  'envior'
        replacement:   '$1'
    - source_labels: ['__meta_consul_service_metadata_application']
        regex:         '(.*)'
        target_label:  'application'
        replacement:   '$1'
    - source_labels: ['__meta_consul_service_metadata_name']
        regex:         '(.*)'
        target_label:  'name'
        replacement:   '$1'
    - source_labels: ['__meta_consul_service_metadata_service']
        regex:         '(.*)'
        target_label:  'service'
        replacement:   '$1'
  - source_labels: ['__meta_consul_service_metadata_Instance']
        regex:         '(.*)'
        target_label:  'Instance'
        replacement:   '$1'
    - source_labels: ['__meta_consul_service']
        regex:         '.*'
        action:     keep  
```

federation configuration:

```yaml
global:
scrape_interval:     20s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
evaluation_interval: 30s # Evaluate rules every 15 seconds. The default is every 1 minute.
scrape_timeout: 20s
external_labels:
  dc: 'global'

scrape_configs:
- job_name: 'federate'
    honor_labels: true
    metrics_path: '/federate'
    params:
    'match[]':
        - '{project=~"ProjectName"}'

static_configs:
  - targets:
    - 'dc1-children-prometheus:9090'
    - 'dc2-children-prometheus:9090'
    - 'dc3-children-prometheus:9090'


- job_name: 'parent-prometheus'
    static_configs:
    - targets: [ 'parent-prometheus' ]

alerting:
alertmanagers:
- static_configs:
    - targets: [ 'AlertManagerList:Port' ]

rule_files:
    - "/data/app/prometheus-2.13.1/conf.d/rules/*.yml"
```

* 安装配置警报管理(alertmanage)

警报安装(alertmanage install):

```bash
cd /data/soft/
wget https://github.com/prometheus/alertmanager/releases/download/v0.18.0/alertmanager-0.18.0.linux-amd64.tar.gz
tar zxf  alertmanager-0.18.0.linux-amd64.tar.gz 
mv alertmanager-0.18.0.linux-amd64 alertmanager-0.18.0
mv  alertmanager-0.18.0 ../app/
```

警报管理(alertmanage configuration):

```bash
cat > /usr/lib/systemd/system/alertmanager.service  <<EOF
[Unit]
Description=Prometheus alertmanager Server
Documentation=https://prometheus.io/docs/introduction/overview/
After=network-online.target

[Service]
User=root
Restart=on-failure
ExecStart=/data/app/alertmanager-0.18.0/bin/alertmanager \
                --config.file=/data/app/alertmanager-0.18.0/conf.d/alertmanager.yml \
                --storage.path=/data/app/alertmanager-0.18.0/data 

ExecReload=/bin/kill -HUP $MAINPID
[Install]
WantedBy=multi-user.target

EOF
```

#### grafana部署安装

* grafana install

```bash
wget https://dl.grafana.com/oss/release/grafana-6.4.4-1.x86_64.rpm
sudo yum localinstall grafana-6.4.4-1.x86_64.rpm
```

* grafana configuration

```ini
#  grep ^[^\;] grafana.ini  | grep ^[^#]
instance_name = server
[paths]
temp_data_lifetime = 12h
logs = /var/log/grafana
plugins = /var/lib/grafana/plugins
provisioning = conf/provisioning
[server]
protocol = http
http_addr = ADDRESS
http_port = PORT
domain = DOMAIN
enforce_domain = true
root_url = %(protocol)s://%(domain)s/
enable_gzip = true
[database]
type = mysql
host = DSN_HOSTNAME
name = DSN_DB_NAME
user = DSN_DB_USERNAME
# 如果密码存在特殊符号，使用 ```PASSWORD```
password = DSN_DB_PASSWORD 
# 如果密码存在特殊符号，使用 base64 转码 base64(password)
url = mysql://DB_USERNAME:DB_PASSWORD@ADDRESS:PORT/DB_NAME

max_idle_conn = 10
max_open_conn = 0
conn_max_lifetime = 10
log_queries = false
[remote_cache]
provider = redis
# redis: config like redis server e.g. `addr=127.0.0.1:6379,pool_size=100,db=grafana`
provider_config = `addr=ADDRESS:PORT,pool_size=500,db=grafana,password=CACHE_PASSWORD`
cookie_name = grafana_sess
cookie_secure = false
session_life_time = 14400
```

#### consul部署安装

* consul install

```bash
wget https://releases.hashicorp.com/consul/1.6.2/consul_1.6.2_linux_amd64.zip
unzip consul_1.6.2_linux_amd64.zip
mv consul /usr/local/bin
```

* consul configuartion
 
systemctl service:

```shell
cat > /etc/systemd/system/consul.service  <<EOF
[Unit]
Description=consul agent
Requires=network-online.target
After=network-online.target

[Service]
EnvironmentFile=-/etc/sysconfig/consul
Environment=GOMAXPROCS=2
Restart=on-failure
ExecStart=/usr/local/bin/consul agent -config-dir=/etc/consul.d/server -rejoin -ui -data-dir=/var/lib/consul
ExecReload=/bin/kill -HUP 
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target

EOF
```



* server configuration
   
```json
{
    "bind_addr": "192.168.50.42",
    "client_addr": "192.168.50.42",
    "bootstrap_expect": 3,
    "server": true,
    "datacenter": "bj2",
    "data_dir": "/var/lib/consul",
    "dns_config": {
        "allow_stale": true,
        "max_stale": "15s"
    },
    "start_join": [
        "192.168.50.41:8301",
        "192.168.50.42:8301",
        "192.168.50.43:8301"
    ],
    "retry_join": [
        "192.168.50.41",
        "192.168.50.42",
        "192.168.50.43",
        "192.168.50.44"
    ],
    "retry_interval": "10s",
    "retry_max": 100,
    "skip_leave_on_interrupt": true,
    "leave_on_terminate": false,
    "ports": {
        "dns": 53,
        "http": 10801
    },
    "rejoin_after_leave": true,
    "addresses": {
        "http": "0.0.0.0",
        "dns": "0.0.0.0"
    }
}
```
client configuration:

```json
{
"bind_addr": "192.168.50.44",
"client_addr": "192.168.50.44",
"bootstrap_expect": 0,
"server": false,
"datacenter": "bj2",
"data_dir": "/var/lib/consul",
"dns_config": {
    "allow_stale": true,
    "max_stale": "15s"
},
"start_join": [
    "192.168.50.41:8301",
    "192.168.50.42:8301",
    "192.168.50.43:8301"
],
"retry_join": [
    "192.168.50.41",
    "192.168.50.42",
    "192.168.50.43",
    "192.168.50.44"
],
"retry_interval": "10s",
"retry_max": 100,
"skip_leave_on_interrupt": true,
"leave_on_terminate": false,
"ports": {
    "dns": 53,
    "http": 80
},
"rejoin_after_leave": true,
"addresses": {
    "http": "0.0.0.0",
    "dns": "0.0.0.0"
  }
}
```

#### consul例子

server discovery example configuration:

```json
{"services": [{
    "id": "idNotEqual",
    "name": "hostname",
    "address": "192.168.6.57",
    "enable_tag_override": true,
    "meta": {
    "role": "master",
    "Instance": "汇总库",
    "name":"as-dsn-conncetion:3306",
    "application":"host",
    "envior":"生产环境",
    "service": "mysql",
    "project":"project"
    },
    "port": 19100
    }]
}
```

grafana显示示例:

1. grafana origin

![grafana.origin.display](/Prometheus/images/ops.grafana.orgin.png)

2. grafana folder
![grafana.folder.display](/Prometheus/images/ops.grafana.folder.png)

3. grafana dashboard
![grafan.dashboard.display](/Prometheus/images/ops.grafana.dashboard.png)



#### Prometheus总结

Prometheus是最初在SoundCloud上构建的开源系统监视和警报工具包。自2012年成立以来，许多公司和组织都采用了Prometheus，该项目拥有非常活跃的开发人员和用户社区。现在，它是一个独立的开源项目，并且独立于任何公司进行维护。
为了强调这一点并阐明项目的治理结构，Prometheus于2016年加入了Cloud Native Computing Foundation，这是继Kubernetes之后的第二个托管项目。

Prometheus 生态圈中包含了多个组件，其中许多组件是可选的：

* Prometheus Server: 用于收集和存储时间序列数据。
* Client Library: 客户端库，为需要监控的服务生成相应的 metrics 并暴露给 Prometheus server。当 Prometheus server 来 pull 时，直接返回实时状态的 metrics。
* Push Gateway: 主要用于短期的 jobs。由于这类 jobs 存在时间较短，可能在 Prometheus 来 pull 之前就消失了。为此，这次 jobs 可以直接向 Prometheus server 端推送它们的 metrics。这种方式主要用于服务层面的 metrics，对于机器层面的 metrices，需要使用 node exporter。
* Exporters: 用于暴露已有的第三方服务的 metrics 给 Prometheus。
* Alertmanager: 从 Prometheus server 端接收到 alerts 后，会进行去除重复数据，分组，并路由到对收的接受方式，发出报警。常见的接收方式有：电子邮件，pagerduty，OpsGenie, webhook 等。


特征:

1. 多维数据模型，Prometheus 将所有的数据都存储为time series,time series由Metric名称和键/值对标识
2. PromQL，Prometheus特有的查询语言，可以充分利用多维数据模型
3. 不依赖分布式存储;单个服务器节点是自治的
4. 时间序列收集通过HTTP上的pull模型进行
5. 通过中间网关支持pushing time series
6. 目标是通过服务发现或静态配置发现的
7. 图形和仪表板支持的多种模式


Prometheus旨在追踪整个系统的健康状况、行为和表现，而不是单个事件

Prometheus架构:

![prometheus architecture](/Prometheus/images/ops.prometheus.Architecture.png)


* Prometheus优点和缺点

Prometheus非常适合记录任何纯数字时间序列。它既适合以机器为中心的监视，也适合监视高度动态的面向服务的体系结构。在微服务世界中，它对多维数据收集和查询的支持是一种特别的优势。

Prometheus的设计旨在提高可靠性，使其成为中断期间要使用的系统，从而使您能够快速诊断问题。每个Prometheus服务器都是独立的，而不依赖于网络存储或其他远程服务。当基础结构的其他部分损坏时，您可以依靠它，而无需建立广泛的基础结构来使用它

Prometheus重视可靠性。即使在故障情况下，您也始终可以查看有关系统的可用统计信息。如果您需要100％的准确性（例如按请求计费），则Prometheus并不是一个不错的选择，因为所收集的数据可能不会足够详细和完整。在这种情况下，最好使用其他系统来收集和分析计费数据，并使用Prometheus进行其余的监视。
