#### project-readme

遇到的问题和需要避免的一些错误的认知,提高自己的能力,我们在路上!

目前在维护的模块包括:

* [Arango](./Arango)
* [Caddy](./Caddy)
* [CockroachDB](./CockroachDB)
* [Cousul](./Consul)
* [Docker](./Docker)
* [Drone](./Drone)
* [Echo](./Echo)
* [ElasticSearch](./ElasticSearch)
* [Encryption](./Encryption)
* [Etcd](./Etcd)
* [Flutter](./Flutter)
* [Gin](./Gin)
* [GitlabCI](/GitlabCI)
* [Gorm](./Gorm)
* [Grafana](./Grafana)
* [Grpc](./Grpc)
* [Istio](./Istio)
* [Jaeger](./Jager)
* [Jenkins](./Jenkins)
* [Kafka](./Kafka)
* [Kong](./Kong)
* [Kubernetes](./Kubernetes)
* [Mqtt](./Mqtt)
* [Mysql](./Mysql)
* [Ngrok](./Ngrok)
* [Nsq](./Nsq)
* [Prometheus](./Prometheus)
* [Rabbitmq](./Rabbitmq)
* [React](./React)
* [Redis](./Redis)
* [Supervisor](./Supervisor)
* [Thrift](./Thrift)
* [TiDB](./TiDB)
* [Traefik](./Traefik)
* [Viper](./Viper)
* [Zipkin](./Zipkin)


#### 维护的基本要求

希望我们每个人在提交的时候先审核下,初步的规范是这样的,一个模块的技术,一个人进行维护,包括使用,安装,遇到的问题,总结等等,在这些内容中可以编辑的
总结成一个文档需要配合图文并茂的展示,然后结合图片显示,放在功能目录asstes下面引用到自己的模块.

git -----> feature-dev ----->master.

git commit -m "feat: XXXX 添加了XXX功能",其中message为提交的具体内容，

注意: feat:<此处有空格>xxxxxx

以下是具体的commit的type类型对应不同的功能:

```markdown
'docs', // 仅仅修改了文档，比如README等
'chore', // 改变构建流程、或者增加依赖库、工具等
'feat', // 添加新功能
'fix', // 修复bug
'refactor', // 代码重构，包括优化相关合代码风格调整
'revert', // 回滚
'test', // 测试用例，包括单元测试、集成测试等
'conf', //配置修改
```

先从本地分支提交代码到feature-dev分支上面,然后审核过了,会合到master的,master是主分支,先不要直接合并到master上面.

谢谢大家的配合和努力,我们一起维护好这个模块,为我们自己学习和提高的同时,也可以帮助其他人学习和提高!

#### project-readme

欢迎大家有兴趣的可以一起贡献,谢谢大家!

License
This is free software distributed under the terms of the MIT license

