应用基础环境变量

---

项目运行需要根据运行环境来决定使用的配置文件、不同项目需要根据应用名称来区分服务器部署路径等等。
这些环境变量需要根据部署环境动态设置，传统方式是靠开发者在上线前进行调整和维护。这样会增加开发者的心智负担，
同时也增加了遗漏的风险，一个比较合理的方式是在发布系统中根据发布的环境手动写入。GO项目所有应用基础环境变量
都集中在golang/lib/application包，便于使用和通过GO连接器进行打入。考虑到本地调试的情况增加本地环境变量的
设置方式，application包里面的变量优先获取连接器打入的值如果没有获取到就尝试从环境变量获取

#### 应用基础环境变量包含：

- vcsInfo
- buildTime
- appName
- appID
- appMode
- logDir

|||
|----|---|
|vcsInfo|由commitID构成的版本号|
|buildTime|打包时间|
|appName|应用全局唯一的英文名称，全公司不会重复|
|appID|应用全局唯一的数字id，全公司不会重复|
|appMode|实例运行环境，local本地开发环境、dev开发联调环境、pre预发布环境、prod线上环境|
|logDir|项目日志记录|

#### 应用基础环境变量本地设置方式：

```
export UNIONTECH_APP_MODE = "local"
export UNIONTECH_APP_ID = "10"
export UNIONTECH_APP_LOG_DIR = "/home/www/logs/"
```

#### 应用基础环境变量使用方法：

```

package main

import (
	"fmt"
	"pkg.deepin.com/golang/lib/application"
)

func main() {
	fmt.Printf("AppName=%s\n", application.AppName())
	fmt.Printf("BuildTime=%s\n", application.BuildTime())
	fmt.Printf("VcsInfo=%s\n", application.VcsInfo())
	fmt.Printf("AppMode=%s\n", application.AppMode())
	fmt.Printf("AppID=%s\n", application.AppID())
	fmt.Printf("LogDir=%s\n", application.LogDir())
}

```

输出：

```

~/src/test$ go run main.go -a

AppName=main
BuildTime=
VcsInfo=
AppMode=local
AppID=0
LogDir=/home/www/logs/main/

```