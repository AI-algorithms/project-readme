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