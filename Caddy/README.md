#### Caddy的使用

```
download: https://caddyserver.com/download

caddy文件解压缩到/usr/bin目录

caddyfile的编写

sub.domainname.com {
    tls your@email.com {
        max_certs 1 //证书最大数量
    }
    proxy / localhost:3000
}
更多语法参考: https://caddyserver.com/docs/caddyfile

在caddyfile目录下运行caddy即可，或在任 意目录运行(ROOT身份),
（目的：输入邮箱生成LETSENCRYPT证书，存放于ROOT根目录.caddy下）

> caddy -conf /path/to/Caddyfile

caddy会下载证书，建立加密通道，并将请求转发至 3000

supervisor守护caddy

如果在caddy.log中出现需要输入邮箱的 prompt则在caddyfile中设定 `tls your@email.com`
supervisorctl.stop caddy
supervisorctl start caddy

> vim /etc/supervisord.d/caddy.ini
or 
> vim /etc/supervisord.conf
supervisor添加配置
[program:caddy]
command=caddy -conf ........./Caddyfile
directory=/root/app/caddy
environment=PWD=/root/app/caddy
autostart=true
autorestart=true
stopsignal=KILL

```
