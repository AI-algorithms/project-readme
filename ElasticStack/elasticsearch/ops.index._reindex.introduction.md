## 关于elastic index _reindex 实施文档

---
- index _reindex 用在那方面

  1.  日志场景下 索引的压缩和合并
  2.  数据迁移和重建
---

### 数据流
![elastic.index._reindex-Architecture](/ElasticStack/images/ops.elastic.index._reindex.png)

---
如何查看es 版本

一般情况下, 通过访问 http 公开访问 elasticsearch 集群。默认情况 http.port 端口为 9200-9300 [http.port](https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-http.html#_settings)

```json
{
  "name": "nodeName",
  "cluster_name": "clusterName",
  "cluster_uuid": "hxtxTmmaSquyzVqpsEir9g",
  "version": {
    "number": "6.5.4",
    "build_flavor": "default",
    "build_type": "tar",
    "build_hash": "d2ef93d",
    "build_date": "2018-12-17T21:17:40.758843Z",
    "build_snapshot": false,
    "lucene_version": "7.5.0",
    "minimum_wire_compatibility_version": "5.6.0",
    "minimum_index_compatibility_version": "5.0.0"
  },
  "tagline": "You Know, for Search"
}
```
---

### elasticsearch 默认情况下支持两种通信模块
- http

    http 模块允许通过 http 公开访问 elasticsearch api, 更多信息可以参考 [elastic http docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-http.html)
- transport

     transport 模块用于集群内节点之间的内部通信, 更多信息可以参考 [elastic transport docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/modules-transport.html)

---
