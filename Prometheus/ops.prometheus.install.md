| time | action|
| :---: | :---: |
|  `2019-11-17 02:45` | prometheus install docs |
||

**业务需求**


**集群规划**

![design.architecture](/Prometheus/images/ops.prometheus.design.architecture.png)

**组件版本**
| name | version| 
| :---: | :---: |
| `prometheus` | 2.13.1|
| `alertmanager` | 0.18.0 
|

**部署安装**

- ***prometheus***

    a. ***instance***
    ```shell
    cat > /etc/systemd/system/prometheus.service <<-EOF
    [Unit]
    Description=Prometheus Server
    Documentation=https://prometheus.io/docs/introduction/overview/
    After=network-online.target

    [Service]
    User=prometheus
    Restart=on-failure
    ExecStart=/usr/local/bin/prometheus-2.13.1/prometheus \\
                                -config.file=/etc/prometheus/prometheus.yml \\
                                -storage.local.path=/var/lib/prometheus/data

    [Install]
    WantedBy=multi-user.target
    EOF
    ```

- ***grafana***

- ***consul***

- ***exporter***


