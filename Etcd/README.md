#### ETCD v2 和 v3 的存储模型和多版本并发控制(MVCC multi-Version Concurrency Control)

etcd `v2` 是一个纯粹的内存数据库, 写操作先通过 Raft 复制日志文件, 复制成功后将数据写入到内存, 整个数据库在内存中是一个简单的树结构, `v2` 并未实时的将数据写入到磁盘, 数据持久化是通过快照来实现的, 具体实现就是讲整个内存中的数据复制一份出来,然后序列化成 JSON ,写入磁盘中成为一个快照, 在 v2 中,每个 key 只保留一个 value , 所以并没有多版本的问题。

在 etcd `v3` 中, 每个 key 的 value 都需要保存多个历史版本, 这就极大的增大了存储的数据量, 而内存容不下这么大的数据量, 所以, 在 v3 版本中, key/value的更改是自然持久化的, 然后需要用户手动或者每隔一段时间定期删除内存中老版本的数据, 这个操作称之为数据的压缩(或删除)。

mvcc的出现只为了解决读写竞争问题, 传统的解决方式是加锁, 加锁也会写写/读写互斥, 虽然这是正常的, 但是在一定的程度上降低了性能, 同时加锁的方式也有一些缺点:

1. 锁的粒度不好控制
2. 读锁和写锁相互阻塞
3. 基于锁的隔离机制会有一段很长的读事务, 在这段时间内这个对象就无法被改写,后面的事务也就被阻塞,直到这个事务完成为止。

所以, mvcc以一种优雅的方式解决了锁带来的问题, 在MVCC中, 每当想要更改和删除某个数据时, DBMS 不会再原地删除和修改这个已有的数据对象本身,而是针对这个数据对象创建一个新的版本, 这样一来, 并发读操作仍然可以读取到老版本的数据, 而写操作也可以同时进行, 这个模式的好处在于可以让读操作不在阻塞,基于版本的情况下, 事实上也不需要锁,Mysql 也使用了mvcc。 

mvcc 在etcd v3版本的实现是: etcd使用了 boltDB数据库来实现, boltDB数据库是一个简单的key/value存储, 且非常轻量, boltDB作为etcd的底层数据库, etcd在BoltDB存储的方式是: 以 reversion 为key(注: reversion ), value是etcd 自己的 key-value组合, 也就说etcd 会在BoltDB中保存每个版本, 从而实现多版本机制。 由于在BoltDB的存储方式是以 reversion 为key, 但是在用户使用的时候是通过key来查询值的, 那么etcd v3是怎么解决的呢? 答案是 etcd v3 在内存中维护了一个kvindex, 保存的就是key与 reversion 之间的关系, 这个kvindex的作用就是用来家属查询的, kvindex 是 B树 实现的, 当用户基于 key 来查询 value 的时候, 会先在kvindex 中查询这个key 对应的所有 reversion, 然后再通过 reversion 从BoltDB中查询数据。etcd v3 实现了mvcc之后, 数据是实时写入到BoltDB数据库的, 这样数据的持久化就分摊到每次对key的写请求上个, 所以,etcd v3 不需要做数据快照, 而etcd v2 需要定时做快照以持久化数据; etcd v3由于基于版本,数据量相对较大, 所以需要定时压缩数据, 米面数据超出磁盘的容量。




#### ETCD v2 和 v3 的事务和隔离

事务必须要满足ACID,即原子性, 一致性, 隔离性, 持久性

在`v2`中, etcd 只提供了针对单个 key 的条件更新操作, 即 CAW(Compare-And-Swap) 操作。 客户端在对一个 key 进行写操作的时候需要提供该key的版本号或者当前值, 服务端会对其版本或者值进行比较, 如果服务端的值和版本已经被更新了,那么 CAW操作会失败。CAW只针对单个key提供了简单信号量和有限的原子操作, 不能满足更新复杂的应用场景, 比如设计到多个key变更事务的时候, etcd v2将无法处理。

在`v3`中,他能够支持更加复杂的事务,比如多个key变更的事务, `v3`使用软件事务内存(STM software Transactional Memory), API则对基于版本号的冲突进行逻辑封装。所谓的STM指的是: 它自动检测内存访问时的冲突, 并尝试在冲突的时候对事物进行回退和重试。etcd v3 的软件事务内存也是乐观的思路控制思路: 在事务最终提交的时候检测是否有冲突, 如果有则回退和重试;而悲观的冲突控制则是在事务开始之前就检测是否有冲突, 如果有就暂不执行。所以,如果冲突比较频繁时,乐观的冲突控制效率就比较差,悲观的冲突控制效率效率反而会更好一些, 因为乐观冲突控制总是在最后一步才检测冲突,导致无效操作比较多。


#### ETCD v2 和 v3 的 watch 机制实现原理和区别

> 原理: 

在`v2`中, watch API是通过HTTP/1.1的long poll实现, 也是一个HTTP GET请求,但和一般的Get不同之处在于,url里会多出一个`?wait=true`参数, etcd v2看到这个参数后, 知道了这是个watch请求, 所以不会立即返回response, 而是知道这个数据被更新后才会返回。服务端的处理逻辑: 当接收到用户的请求后, Server 会调用 store 的watch()方法, store 的 watch() 方法会调用 WatcherHub 的watch() 方法, 然后将watch 请求的信息添加到 WatcherHub 里面, 然后当 etcd 的 key 更新时, 就会生成一个 Event 事件, 然后调用 WatchHub 的notify() 方法通知所有正在 watch 该 key 的 Watcher(即Client).

在`v3`中, etcd会保存每个客户端发过来的watch请求, watch被细分为两种, watch 单个key和 watch 前缀, 所以, watchGroup 包含了两种 Watcher: `Key Watchers` 和 `range Watchers` , Key Watchers的数据结构是 每个key对应一组 Watcher, range Watchers的数据结构是一个线段树, 可以方便地通过区间查找到对应的 Watcher。 etcd 有一个线程会持续不断地遍历所有的 watch 请求, 每个 watch 对象都会负责维护其监控的 key 事件, 看将其推送到具体某revision。

etcd 会根据这个revision的main ID去bblot 中继续向后遍历, 而bblot是一个按照key有序排列的key-value引擎, 而bblot中的key是由revision 由revision.main ID + revision.sub ID组成的, 所以遍历就会依次经过历史上发送过的所有事务(tx)的记录, 遍历经历过的每个K-V, etcd都会反序列化其中的value,也就是mvccpb.KeyValue,判断其中的key是否是watch请求关注的, 如果是就发送给客户端。 其中遍历有个优化的方式就是,并不是每一个的Watcher单独遍历bblot从中找到属于自己关注的key, 因为这样性能很差, 而是在遍历bblot的时候, JSON会反序列化每个mvccpb.KeyValue结构, 判断其中的key是否属于watchGroup关注的key,是的话就发送给客户端。

> 区别:

1. 连接方式不同 

在 `v2` 中,watch API是基于HTTP的long poll实现的, 本质就是一个 HTTP1.1 的长连接, 因此一个watch请求需要维护一个TCP连接, 假设某个客户端分别watch了10个key, 那么就同一个客户端需要创建10个TCP连接,所以,服务端就需要耗费资源用于维持TCP连接。

在 `v3` 中, etcd使用了gRPC, 而gRPC又利用了http/2的TCP链接多路复用, 这样同一个客户端的不同watch可以共用一个TCP链接,这样就极大的减少了每个watch所带来的资源消耗。

2. 历史记录限制的问题

在 `v2` 中, 由于服务器端只保留了最新的1000个记录,所以,v2版本的watch机制只能watch最近1000条记录, 因此很难通过watch机制来实现完整的数据同步(有数据丢失的可能)；

在 `v3` 中, 只要历史版本没有被压缩, 那么这个数据就能被watch到。

3. watch key的方式不同

在`v2`中, watch机制只能watch单个key

在`v3`中, watch除了支持单个 key watch, 还支持前缀 key watch





