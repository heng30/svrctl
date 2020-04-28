# svrctl 是 golang 编写的简单进程起停控制

## 主要功能：
- start: 启动程序，将pid写入/var/run/program_name.pid 文件
- stop: 发送SIGINT信号给正在运行的程序，退出程序
- restart: 重新启动程序
- daemon: 以后台方式启动程序



## 注意：
程序会消耗掉 start, stop, restart, daemon 这四个参数



## 例子
```go
package main

import (
	"github.com/Heng30/svrctl"

	"time"
)

func main() {
	svrctl.Run()

	for true {
		time.Sleep(time.Second)
	}
}
```
