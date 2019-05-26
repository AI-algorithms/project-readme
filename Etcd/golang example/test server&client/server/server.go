package main

import (
	"cfgcenter"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	serverID   = 0
	serverName = "unknown"
	serverType = "/dc/"
	serverIP   = "127.0.0.1"
	serverPort = 0
)

/*
	自动动态配置与注册
	对外提供服务
*/

func main() {
	if len(os.Args) < 3 {
		panic("usage: " + os.Args[0] + " serverID serverName")
	}

	//1,2,3...
	serverID = func() int {
		n, err := strconv.Atoi(os.Args[1])
		if err != nil {
			panic("serverID has err: " + err.Error())
		}
		return n
	}()
	//beijing,shanghai ...
	serverName = os.Args[2]

	//链接服务注册中心
	if err := cfgcenter.Connect([]string{"127.0.0.1:2379"}, "", "", 0); err != nil {
		log.Fatalln(err)
	}
	defer cfgcenter.Close()

	//对外提供服务,server port 随机
	listener, err := net.Listen("tcp", serverIP+":"+strconv.Itoa(serverPort))
	if err != nil {
		log.Fatalln(err)
	}
	//获取server port
	serverPort, err = strconv.Atoi(strings.Split(listener.Addr().String(), ":")[1])
	if err != nil {
		log.Fatalln(err)
	}

	//对外处理服务,要在注册前确保accpet
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("err:", err)
				continue
			}

			go handle(conn) //简单的返回时间和服务名称
		}
	}()

	fmt.Println("listen on ", serverIP, serverPort)

	//注册自己到服务中心
	if err := cfgcenter.Register(serverType+serverName, serverID, map[string]interface{}{"ip": serverIP, "port": serverPort, "name": serverName, "type": serverType, "id": serverID}); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Register complete!")

	//处理某些定时任务等...
	select {}
}

//handle 简单返回时间和服务信息
func handle(c net.Conn) {
	defer c.Close()
	_, err := c.Write([]byte(fmt.Sprintf("%v,name:%s,type:%s,id:%d", time.Now(), serverName, serverType, serverID)))
	if err != nil {
		log.Println("handle has err:", err)
	} else {
		log.Println("handle", c.RemoteAddr())
	}
}
