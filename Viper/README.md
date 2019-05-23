## 安装
go get github.com/spf13/viper

## ECTD集成(GPG加密)

```
$> go get github.com/xordataexchange/crypt
$> go install github.com/xordataexchange/crypt/bin/crypt
$> vim app.batch
```

```
%echo Generating a configuration OpenPGP key
Key-Type: default
Subkey-Type: default
Name-Real: app
Name-Comment: app configuration key
Name-Email: app@example.com
Expire-Date: 0
%pubring .pubring.gpg
%secring .secring.gpg
%commit
%echo done
```

```
$> gpg2 --batch --armor --gen-key app.batch
$> crypt set -keyring .pubring.gpg /config/config.toml config.toml 
$> docker run -d -p 4001:4001 --name some-etcd elcolio/etcd:latest
```

**go sample code**

```
package main

import (
        "fmt"

        "github.com/spf13/viper"
        _ "github.com/spf13/viper/remote"
)

func main() {
        viper.AddSecureRemoteProvider("etcd","http://127.0.0.1:4001","/config/config.toml",".secring.gpg")
        viper.SetConfigType("toml")
        go func(){viper.WatchRemoteConfig()}()
        err := viper.ReadRemoteConfig()
        if err != nil {
                fmt.Println(err)
                return
        }
        fmt.Println(viper.GetString("postgres.port"))
        var w string
        fmt.Scanf("%s", &w)
        fmt.Println(w)
        fmt.Println(viper.GetString("postgres.port"))
}
```
