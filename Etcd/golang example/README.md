### 说明

- **cfgcenter** 文件夹保存了服务配置中心的代码,etcd的api在这里调用
- **test server&client** 文件夹保存了一个服务端和一个客户端,通过调用**cfgcenter** 中的API实现了简单的服务注册与发现
- etcd必须监听`127.0.0.1:2379`,如果不是该地址,需要全局搜一下`127.0.0.1:2379`字符串,修改链接地址

#### 使用

1. 进入server目录 `cd test server&client/server`
2. 编译`go build `
3. 在两个终端**分别运行**两个server端程序 `./server 1 beijing`  `./server 2 shanghai`
4. 进入client目录 `cd test server&client/client
5. 编译`go build `
6. 运行client程序 `./client`
7. 此时看到client会**自动链接**到server端,每隔5s调用一次
8. 此时停止一个server端,可以看到client端在**短时间**内(心跳间隔默认10s)会尝试调用已经离线的server端,但最终会自动将离线的server端剔除可用服务列表,并将调用分发到其他可用服务上
9. 此时启动一个server端,可以看到client端**立即**识别了该服务,并将其添加到可用服务列表中,同时会将调用平均分发

