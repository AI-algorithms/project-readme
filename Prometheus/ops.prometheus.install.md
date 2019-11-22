| time | action|
| :---: | :---: |
|  `2019-11-17 02:45` | prometheus install docs |


**业务需求**

- ***多维度满足业务项目的需求***
- ***统一监控系统***
- ***满足云原生需求***

**集群规划**

![design.architecture](/Prometheus/images/ops.prometheus.design.architecture.png)

**组件版本**

| name | version|  path | 
| :---: | :---: | :---: |
| `prometheus` | 2.13.1| /data/app/prometheus-2.13.1 |
| `alertmanager` | 0.18.0 | /data/app/alertmanager-0.18.0 |
| `grafana` | 6.3.3 |  rpm |
| `consul` | 1.6.2 | /usr/local/bin | 


**部署安装**

**prometheus**

- ***prometheus install***

    ```shell
    cd /data/soft
    wget https://github.com/prometheus/prometheus/releases/download/v2.13.1/prometheus-2.13.1.linux-amd64.tar.gz

    tar xzf prometheus-2.13.1.linux-amd64.tar.gz
    mv prometheus-2.13.1.linux-amd64 prometheus-2.13.1
    mv prometheus-2.13.1 ../app/
    ```

- ***instance systemctl service***

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

- ***instance configuration***

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