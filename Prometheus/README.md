### project-readme

遇到的问题和需要避免的一些错误的认知,提高自己的能力,我们在路上!


#### prometheus结合grafana使用

| time | action|
| :---: | :---: |
|  `2019-11-17 02:45` | prometheus install docs |


业务需求:

* 多维度满足业务项目的需求
* 统一监控系统
* 满足云原生需求

集群规划:
[design.architecture](/Prometheus/images/ops.prometheus.design.architecture.png)

组件版本:

| name | version|  path | 
| :---: | :---: | :---: |
| `prometheus` | 2.13.1| /data/app/prometheus-2.13.1 |
| `alertmanager` | 0.18.0 | /data/app/alertmanager-0.18.0 |
| `grafana` | 6.3.3 |  rpm |
| `consul` | 1.6.2 | /usr/local/bin | 


#### 部署安装

* prometheus

```shell
cd /data/soft
wget https://github.com/prometheus/prometheus/releases/download/v2.13.1/prometheus-2.13.1.linux-amd64.tar.gz

tar xzf prometheus-2.13.1.linux-amd64.tar.gz
mv prometheus-2.13.1.linux-amd64 prometheus-2.13.1
mv prometheus-2.13.1 ../app/
```

* instance systemctl service

```shell
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

* instance configuration

    `children instance configuartion`
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

    `federation configuration`
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

**alertmanage**

- ***alertmanage install***

    ```shell
    cd /data/soft/
    wget https://github.com/prometheus/alertmanager/releases/download/v0.18.0/alertmanager-0.18.0.linux-amd64.tar.gz
    tar zxf  alertmanager-0.18.0.linux-amd64.tar.gz 
    mv alertmanager-0.18.0.linux-amd64 alertmanager-0.18.0
    mv  alertmanager-0.18.0 ../app/
    ```
- ***alertmanage configuration***

    ```shell
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

**grafana**

- ***grafana install***

    ```shell
    wget https://dl.grafana.com/oss/release/grafana-6.4.4-1.x86_64.rpm
    sudo yum localinstall grafana-6.4.4-1.x86_64.rpm
    ```

- ***grafana configuration***

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

**consul**

- ***consul install***

    ```shell
    wget https://releases.hashicorp.com/consul/1.6.2/consul_1.6.2_linux_amd64.zip
    unzip consul_1.6.2_linux_amd64.zip
    mv consul /usr/local/bin
    ```

- ***consul configuartion***
 
    `systemctl service`
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

    `server configuration`
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

    `client configuration`
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


**example**

- ***server discovery example configuration***

    ```json
    {
    "services": [
        {
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
            }
        ]
    }
    ```


- ***grafana display example***

`grafana origin`
![grafana.origin.display](/Prometheus/images/ops.grafana.orgin.png)

`grafana folder`
![grafana.folder.display](/Prometheus/images/ops.grafana.folder.png)

`grafana dashboard`
![grafan.dashboard.display](/Prometheus/images/ops.grafana.dashboard.png)


#### 概览

| time | action |
| :---: | :----: |
| `2019-11-17 02:20` | `prometheus monitor` |
||

**OVERVIEW**

    Prometheus is an open-source systems monitoring and alerting toolkit originally built at SoundCloud. Since its inception in 2012, many companies and organizations have adopted Prometheus, and the project has a very active developer and user community. It is now a standalone open source project and maintained independently of any company. To emphasize this, and to clarify the project's governance structure, Prometheus joined the Cloud Native Computing Foundation in 2016 as the second hosted project, after Kubernetes.

**Features**

    a. a multi-dimensional data model with time series data identified by metric name and key/value pairs
    b. PromQL, a flexible query language to leverage this dimensionality
    c. no reliance on distributed storage; single server nodes are autonomous
    d. time series collection happens via a pull model over HTTP
    e. pushing time series is supported via an intermediary gateway
    f. targets are discovered via service discovery or static configuration
    g. multiple modes of graphing and dashboarding support

**Components**

    a. the main Prometheus server which scrapes and stores time series data
    b. client libraries for instrumenting application code
    c. a push gateway for supporting short-lived jobs
    d. special-purpose exporters for services like HAProxy, StatsD, Graphite, etc.
    e. an alertmanager to handle alerts
    f. various support tools

**Architecture**

![prometheus architecture](/Prometheus/images/ops.prometheus.Architecture.png)



**When does it fit?**

    Prometheus非常适合记录任何纯数字时间序列。它既适合以机器为中心的监视，也适合监视高度动态的面向服务的体系结构。在微服务世界中，它对多维数据收集和查询的支持是一种特别的优势。

    Prometheus的设计旨在提高可靠性，使其成为中断期间要使用的系统，从而使您能够快速诊断问题。每个Prometheus服务器都是独立的，而不依赖于网络存储或其他远程服务。当基础结构的其他部分损坏时，您可以依靠它，而无需建立广泛的基础结构来使用它

**When does it not fit?**


    普罗米修斯重视可靠性。即使在故障情况下，您也始终可以查看有关系统的可用统计信息。如果您需要100％的准确性（例如按请求计费），则Prometheus并不是一个不错的选择，因为所收集的数据可能不会足够详细和完整。在这种情况下，最好使用其他系统来收集和分析计费数据，并使用Prometheus进行其余的监视。
