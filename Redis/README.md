#### Redis

Redis 是一个开源（BSD许可）的，内存中的数据结构存储系统，它可以用作数据库、缓存和消息中间件。
 
它支持多种类型的数据结构，如 字符串（strings）， 散列（hashes）， 列表（lists）， 集合（sets）， 有序集合（sorted sets） 与范围查询， bitmaps， hyperloglogs 和 地理空间（geospatial） 索引半径查询。
 
Redis 内置了 复制（replication），LUA脚本（Lua scripting）， LRU驱动事件（LRU eviction），事务（transactions） 和不同级别的 磁盘持久化（persistence）， 并通过 Redis哨兵（Sentinel）和自动 分区（Cluster）提供高可用性（high availability）。

---

主要会包括以下几个方面：

1. Redis使用，安装和配置选项
2. Redis源码进阶
3. Redis集群方案 
4. Redis使用场景

现有目录：

- [安装及配置](./0.1.md)
- [String结构](./0.2.md)
- [Hash结构](./0.3.md)
- [List结构](./0.4.md)
- [Int set结构](./0.15.md)
- [Ziplist 结构](./0.5.md)
- [redisObject结构](./0.6.md)
- [事件驱动](./0.14.md)
- [集群方案](./0.7.md)
- [使用场景](./0.8.md)
- [订阅和发布](./0.9.md)
- [持久化方式](./0.10.md)
- [工作模式](./0.11.md)
- [事务](./0.12.md)
- [通信协议](./0.13.md)