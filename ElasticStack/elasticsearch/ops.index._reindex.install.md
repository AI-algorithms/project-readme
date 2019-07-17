***操作流程***

| name | version |
| :----: | :----: |
 kibana | 6.5.4 
 elasticsearch | 6.5.4

## 以下操作都是基于 kibana dev tools 的工具
## 基于 Index 名称为 ops-container 做演示
- setting dest index
- run _reindex
- check task
- update dest index 
- append alias to dest index
- remove src index alias
- _close src index
---

检查 迁移后的索引对应的模板是否存在且字段是否已设置正确,在测试过程中, es dynamic-mapping 不关闭.

```yaml
GET _template/ops-container-template
```
![get.index.template](/ElasticStack/images/ops.elastic.index.get.index.template.png)

创建迁移后的索引名称并设置副本集和刷新间隔
```yaml
PUT ops-container-demo-display
{
    "index" : {
        "refresh_interval" : "-1",
        "number_of_replicas" : 0
    }
}
```
![put.index.setting](/ElasticStack/images/ops.elastic.index.put.index.setting.png)

执行 _reindex 迁移索引数据, 因为迁移的数据量太大, 添加 wait_for_completion=false 让当前的 _reindex 响应 taskid 
```yaml
# source.index 是 slice 类型
# source.size 默认是 1000
POST _reindex?wait_for_completion=false
{
  "conflicts": "proceed",
  "source": {
    "index": ["ops-container-2019.06","ops-container-2019.05", "ops-container-2019.04"],
    "size": 1000
  },
  "dest": {
    "index": "ops-container-demo-display",
    "op_type": "create"
  }
}

# 运行 task api查看当前 index _reindex 的进度
GET _tasks?detailed=true&actions=*reindex
```
![post.index._reindex.run](/ElasticStack/images/ops.elastic.index.post.index._reindex.run.png)

检查 _reindex 后 迁移后index 跟 原迁移 index 是否一致。
1. 检查 Index 文档数是否一致
2. 检查 _reindex 期间 es 运行日志是否有对应的_reindex 报错日志

```yaml
# 165025
GET ops-container-2019.06/_count
# 7174 
GET ops-container-2019.05/_count
# 4066
GET ops-container-2019.04/_count
# 176265
GET ops-container-demo-display/_count
```
![check.index._count](/ElasticStack/images/ops.elastic.index.check.index._count.png)

更新迁移后的index 的副本集
```yaml

PUT ops-container-demo-display/_settings 
{
    "index" : {
        "refresh_interval" : "15s",
        "number_of_replicas" : 1
    }
}

# 验证设置是否生效
GET ops-container-demo-display/_settings 
```
![check.index._settings_](/ElasticStack/images/ops.elastic.put.index._settings.png)

ops.elastic.put.index._settings.png
别名转换及关闭index
```yaml
POST _aliases
{
  "actions": [
    {
      "add": {
        "index": "ops-container-demo-display",
        "alias": "demo-container"
      }
    }, {
      "remove": {
        "index": "ops-container-2019.06",
        "alias": "demo-container"
      }
    },{
      "remove": {
        "index": "ops-container-2019.05",
        "alias": "demo-container"
      }
    },{
      "remove": {
        "index": "ops-container-2019.04",
        "alias": "demo-container"
      }
    }
  ]
}

POST ops-container-2019.06/_close
POST ops-container-2019.05/_close
POST ops-container-2019.04/_close

```
![alise.index](/ElasticStack/images/ops.elastic.index.post.index._aliases.add.png)
![close.index](/ElasticStack/images/ops.elastic.index.post.index._close.png)