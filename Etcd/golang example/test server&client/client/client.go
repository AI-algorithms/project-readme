package main

import (
	"cfgcenter"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

var (
	backServerType = "/dc/"
)

/*
	自动动态获取server后端服务列表
	定时调用服务,模拟业务处理
	省略了client注册到服务中心的过程
*/
//后端服务列表
var backSrv struct {
	sync.Mutex
	addrs     map[string]*net.TCPAddr
	available bool
}

func init() {
	backSrv.addrs = make(map[string]*net.TCPAddr, 1)
	rand.Seed(time.Now().Unix())
}

func main() {
	//链接服务注册中心
	if err := cfgcenter.Connect([]string{"127.0.0.1:2379"}, "", "", 0); err != nil {
		log.Fatalln(err)
	}
	defer cfgcenter.Close()

	//获取后端服务列表,并保持动态更新
	wchan := make(chan cfgcenter.WatchChg, 5)
	cancle, err := cfgcenter.Watch(backServerType, wchan)
	if err != nil {
		log.Fatalln(err)
	}
	defer cancle()
	//动态更新
	go func() {
		for chg := range wchan {
			if chg.Event == cfgcenter.Online { //服务上线
				backSrv.Lock()
				a, err := net.ResolveTCPAddr("tcp", chg.Infos["ip"]+":"+chg.Infos["port"])
				if err != nil {
					log.Println("ResolveTCPAddr", err)
				} else {
					backSrv.addrs[chg.Key] = a
					backSrv.available = true
					log.Println("update backSrv", backSrv.addrs)
				}
				backSrv.Unlock()
			} else { //服务下线
				backSrv.Lock()
				delete(backSrv.addrs, chg.Key)
				if len(backSrv.addrs) < 1 {
					backSrv.available = false
				}
				log.Println("update backSrv", backSrv.addrs)
				backSrv.Unlock()
			}
		}
	}()

	//模拟调用服务
	for range time.Tick(5 * time.Second) {
		if !backSrv.available {
			continue
		}

		backSrv.Lock()
		//实测靠map的随机不能达到负载均衡,怀疑map随机为了效率舍弃了公平性
		// for _, v := range backSrv.addrs {
		// 	if v != nil {
		// 		go call(v)
		// 		break
		// 	}
		// }
		mapKeys := make([]string, 0, len(backSrv.addrs))
		for k, _ := range backSrv.addrs {
			mapKeys = append(mapKeys, k)
		}
		randK := mapKeys[rand.Intn(len(mapKeys))]
		go call(backSrv.addrs[randK])
		backSrv.Unlock()
	}
}

func call(addr *net.TCPAddr) {
	conn, err := net.DialTCP(addr.Network(), nil, addr)
	if err != nil {
		log.Println("DialTCP", err)
		return
	}
	defer conn.Close()

	var msg [1024 * 10]byte
	for {
		if _, err = conn.Read(msg[:]); err != nil {
			if err != io.EOF {
				log.Println("tcp.Read", err)
			}
			break
		}
		log.Println("call server:", addr.String(), "msg:", string(msg[:]))
	}
}
