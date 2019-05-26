package cfgcenter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

type ConfigCenter struct {
	etcd       *clientv3.Client
	isRegister bool
}

var conn ConfigCenter

const (
	keepAliveTime = 10 //time.Second
)

//Connect 初始化etcd链接
func Connect(endpoints []string, uname, pwd string, timeout time.Duration) error {
	var err error
	conn.etcd, err = clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
		Username:    uname,
		Password:    pwd,
	})

	if err != nil {
		return err
	}
	//...做一些特殊处理

	return nil
}

//Close 关闭etcd链接
func Close() {
	if conn.etcd != nil {
		conn.etcd.Close()
	}
}

//Register 将服务注册到etcd并保持租约(心跳)
func Register(name string, serverID int, infos map[string]interface{}) error {
	if conn.isRegister {
		//update or return error
		return errors.New("registered")
	}
	//设置keep秒租约（过期时间）
	leaseResp, err := conn.etcd.Lease.Grant(context.Background(), keepAliveTime)
	if err != nil {
		return err
	}
	//拿到租约id
	leaseID := leaseResp.ID

	//设置一个ctx取消自动续租
	ctx, cancleFunc := context.WithCancel(context.Background())
	//开启自动续租
	keepChan, err := conn.etcd.Lease.KeepAlive(ctx, leaseID)
	if err != nil {
		cancleFunc()
		return err
	}

	//KeepAlive respond需要丢弃,否则会在标准错误打警告信息
	go func() {
		for range keepChan {
			// eat messages until keep alive channel closes
		}
	}()

	b, err := json.Marshal(infos)
	if err != nil {
		cancleFunc()
		return err
	}

	//直接覆盖
	_, err = conn.etcd.Put(context.Background(), name+"."+strconv.Itoa(serverID), string(b), clientv3.WithLease(leaseID))
	if err != nil {
		cancleFunc()
		return err
	}

	conn.isRegister = true
	return nil
}

//WatchChg 传出变化消息,实际使用中通常与后端服务列表紧耦合,合起来对外提供服务发现功能
type WatchChg struct {
	Event ServerState //Online or Offline
	Key   string
	Infos map[string]string
}

//服务状态
type ServerState int

const (
	//Online 服务转为上线状态
	Online ServerState = iota
	//Offline 服务转为离线状态
	Offline
)

//Watch 实时更新所有可用后端服务(serverType)的实例
func Watch(serverType string, c chan WatchChg) (func(), error) {
	rangeResp, err := conn.etcd.Get(context.Background(), serverType, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	curRevision := rangeResp.Header.GetRevision()
	// 监听后续的PUT与DELETE事件,cancel用于在外部取消,没需求可以不返回
	ctx, cancel := context.WithCancel(context.Background())
	watchChan := conn.etcd.Watch(ctx, serverType, clientv3.WithPrefix(), clientv3.WithRev(curRevision))

	// 获取现有的
	go func() {
		for _, kv := range rangeResp.Kvs {
			c <- creaetWatchChg(clientv3.EventTypePut, kv.Key, kv.Value)
		}
	}()

	// 监听变化
	go func() {
		for watchResp := range watchChan {
			if watchResp.Canceled || watchResp.Err() != nil {
				//被关闭或发生错误
				//log.Errorln(...)
				break
			}

			for _, event := range watchResp.Events {
				c <- creaetWatchChg(event.Type, event.Kv.Key, event.Kv.Value)
			}
		}
	}()

	return cancel, nil
}

func creaetWatchChg(e mvccpb.Event_EventType, key, val []byte) (w WatchChg) {
	w.Key = string(key)
	switch e {
	case clientv3.EventTypePut:
		//服务上线
		w.Event = Online
	case clientv3.EventTypeDelete:
		//服务下线
		w.Event = Offline
		return
	}

	//服务上线或更新
	var valMap map[string]interface{}
	if err := json.Unmarshal(val, &valMap); err != nil {
		log.Println("json err", err)
		return
	}

	w.Infos = make(map[string]string, len(val))
	for k, v := range valMap {
		w.Infos[k] = fmt.Sprint(v)
	}
	return
}

//Lock 获取分布式锁(阻塞超时方式)
//key		锁的key
//timeout	获取锁的最长等待时间,0为永久阻塞等待
//return 	func():用于unlock. 服务获取锁后挂掉,默认60s后会释放
func Lock(key string, timeout time.Duration) (func(), error) {
	s, err := concurrency.NewSession(conn.etcd)
	if err != nil {
		return nil, err
	}
	m := concurrency.NewMutex(s, key)

	var ctx context.Context
	if timeout == 0*time.Second {
		//阻塞抢锁
		ctx = context.Background()
	} else {
		//超时等待抢锁
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), timeout) //设置超时
		defer cancel()
	}

	if err = m.Lock(ctx); err != nil {
		if err == context.DeadlineExceeded {
			//超时...
		}
		return nil, err
	}

	return func() {
		m.Unlock(context.Background())
	}, nil
}
