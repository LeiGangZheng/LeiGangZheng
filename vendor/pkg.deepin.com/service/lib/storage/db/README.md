db基于gorm ([github.com/jinzhu/gorm]()) 封装，支持一主多从、JSON格式日志，慢日志提示。

#### 安装

------

```
go get pkg.deepin.com/service/lib/storage/db
```

#### 快速开始

------

##### 单数据库

```
package main

import (
	"pkg.deepin.com/service/lib/storage/db"
	"pkg.deepin.com/service/lib/log"
	"pkg.deepin.com/service/utils/config"
)

func main() {
	// 加载配置文件
	type MyfConfig struct {
		DB *db.ConnConf // DB配置
	}
	var cfg MyfConfig
	err := config.Load("db.toml", &cfg)
	if nil != err {
		panic(err)
	}
	
	// 初始化默认的DB链接
	db.InitConn(cfg.DB)

	type User struct {
		Id   int
		Age  int
		Name string
	}
	
	var user User
	db.Default().First(&user, "name = ?", "lsw")      // 读取
	db.Default().Create(&User{Id: 1000, Name: "lsw"}) // 创建
	db.Default().Model(&user).Update("Age", 999)      // 更新
	db.Default().Delete(&user)                        // 删除
}

```

###### 配置文件格式

```
[DB]                            
	dialect = "mysql"   # DB类型:mysql
	dsn = "root:123456@tcp(10.0.10.27:3301)/test?parseTime=true&loc=Local&charset=utf8mb4,utf8"      # 地址
	maxOpenConns = 100                      # 最大连接数
	maxIdleConns = 50                       # 最大空闲连接数
	connMaxLifetime = 300                   # 最大空闲回收时间，单位：s
	logMode = true                          # 是否开启debug日志
	logSQL =true                            # 是否显示sql中具体值
	slowThreshold = 500                     # 慢日志阈值，单位：ms
			
```

##### 一主多从

```
package main

import (
	"pkg.deepin.com/service/lib/storage/db"
	"pkg.deepin.com/service/lib/log"
	"pkg.deepin.com/service/utils/config"
)

func main() {
	// 加载配置文件
	type MyfConfig struct {
		DB *db.Conf // DB Manager配置
	}
	var cfg MyfConfig
	err := config.Load("db.toml", &cfg)
	if nil != err {
		panic(err)
	}
	
	// 初始化默认的DBManager
	db.InitManager(cfg.DB)

	type User struct {
		Id   int
		Age  int
		Name string
	}
	
	var user User
	db.Master("test").First(&user, "name = ?", "lsw")      // 从Master读取
	db.Master("test").Create(&User{Id: 1000, Name: "lsw"}) // 从Master创建
	db.Master("test").Model(&user).Update("Age", 999)      // 从Master更新
	db.Master("test").Delete(&user)                        // 从Master删除
	db.Slave("test").First(&user, "name = ?", "lsw")       // 从Slave读取
}

```

###### 配置文件格式

```
[DB]                            
	[DB.test]                                       # DB组名称:test
		[DB.test.Master]                            # Master配置
			dialect = "mysql"                       # DB类型:mysql
			dsn = "root:123456@tcp(10.0.10.27:3301)/test?  parseTime=true&loc=Local&charset=utf8mb4,utf8"      # 地址
			maxOpenConns = 100                      # 最大连接数
			maxIdleConns = 50                       # 最大空闲连接数
			connMaxLifetime = 300                   # 最大空闲回收时间，单位：s
			logMode = true                          # 是否开启debug日志
			logSQL =true                            # 是否显示sql中具体值
			slowThreshold = 500                     # 慢日志阈值，单位：ms
		[DB.test.Slave]                             # Slave配置
			dialect = "mysql"
			dsn = ["root:123456@tcp(10.0.10.27:3302)/test?parseTime=true&loc=Local&charset=utf8mb4,utf8", "root:123456@tcp(10.0.10.27:3301)/test?parseTime=true&loc=Local&charset=utf8mb4,utf8"]
			maxOpenConns = 100
			maxIdleConns = 50
			connMaxLifetime = 300
			logMode = true
			logSQL =true
			slowThreshold = 500
			
```

#### 创建Conn

函数式选项模式创建单个数据库连接。

```
conn, err :=  db.NewConn(
	db.WithDialect(dialect),
	db.WithDSN(dsn),
	db.WithMaxOpenConns(maxOpenConns),
	db.WithMaxIdleConns(maxIdleConns),
	db.WithConnMaxLifetime(connMaxLifetime),
	db.WithLogMode(logMode),
	db.WithLogSQL(logSQL),
	db.WithSlowThreshold(slowThreshold）,
)
```

- WithDialect(dialect string)：设置db类型，默认值为mysql，可取值为mysql和postgres。
- WithDSN(dsn string)：设置db地址。
- WithMaxOpenConns(maxOpenConns int)：设置最大连接数，默认值为100。
- WithMaxIdleConns(maxIdleConns int)：设置最大空闲连接数，默认值10。
- WithConnMaxLifetime(maxIdleConns string)：设置空闲回收时间，默认值300s。
- WithLogMode(logMode bool)：设置是否打印debug日志，默认值为false。
- WithLogSQL(logSQL bool)：是否显示sql中具体值，默认值为true。
- WithSlowThreshold(slowThreshold int)：设置慢日志阈值，默认值为500ms。

#### 创建Group

用于一主多从。

```
NewGroup(groupConf *GroupConf, policies ...GroupPolicy) (group *Group, err error)
```

其中policies为选择slave库的策略，目前主要实现了以下三种策略：

- RandomPolicy()： 随机选择Slave，默认策略。
- RoundRobinPolicy()：轮询选择Slave。
- WeightRandomPolicy(weights []int)：根据权重选择Slave，weights的长度依据Slave的长度而定。

此外，可参考GroupPolicy接口进行自定义策略。

```
dbGroup, _ := db.NewGroup(groupConf, db.RandomPolicy())
dbGroup.Raw("select * from users;").Scan(&users)          // 默认从Master读
dbGroup.Slave().Raw("select * from users;").Scan(&users)  // 根据略选择Slave读
```

#### 创建Manager 

用于多个Group，通过名称获取Group。

```
NewManager(config *Conf) (dbMgr *Manager, err error)
```

```
dbMgr, err := db.NewManager(config)               // 创建Manager 
dbGroup, err := dbMgr.Group(dbName)               // 获取Group
```

#### 日志说明

db的日志等级分为debug、warn、error级别，慢日志的级别为warn，日志示例如下：

```
// error
{	
	"level":"error",
	"time":"2020-07-02 13:59:39",
	"caller":"db@v0.0.0-20200630060405-caccea5b6213/conn.go:90",
	"msg":"error db log",
	"db_caller":"/home/liaoshiwei/Desktop/db/testbeauty/beauty/internal/example/api/handler/user.go:27",
	"sql":"INSERT INTO `users` (`id`,`name`,`age`) VALUES (1000,test,28)",
	"cost":"734.483µs",
	"tableName":"users",
	"dbName":"test",
	"addr":"10.0.10.27:3302",
	"rowsAffected":0,
	"err":"Error 1062: Duplicate entry '1000' for key 'PRIMARY'"
}
```

- level：日志等级。
- time：日志打印时间。
- caller：调用log库的代码所在处。
- db_caller：调用db库的代码所在处。
- sql：实际执行的sql语句。
- cost：执行sql的耗时。
- tableName：表名。
- dbName：数据库名。
- addr：数据库地址。
- rowsAffected：执行sql返回或影响的行数。
- err：具体错误信息。

#### 批量插入

------

```
err := db.Default().Create([]*User{
    &User{Id: 11111, Name: "11111"},
    &User{Id: 22222, Name: "2222"},
}).Error
```

更多详细用法可参考：https://gitlab.deepin.io/golang/lib/gorm。

#### 扩展方法

------

- func (conn *Conn) AutoPreload(auto bool) *Conn

  AutoPreload 设置自动预加载。

  ```
  result:= make([]User, 0)
  conn.AutoPreload(true).Find(&result)
  ```
  
- func (conn *Conn) ForUpdate() *Conn

  ForUpdate 添加数据库排他锁。

  ```
  var result User
  conn.ForUpdate().First(&result, id)
  ```
  
- func (conn *Conn) GetOne(result interface{}, query interface{}, args ...interface{}) (isExist bool, err error)

  GetOne 根据条件查询第一条记录，并返回是否存在。

  ```
  var result User
  query := User{Id: 1002}
  isExist, err := conn.GetOne(&result, query)
  ```

- func (conn *Conn) GetList(result interface{}, query interface{}, args ...interface{}) error

  GetList 根据条件查询记录。

  ```
  query := User{
      Id: 1002,
  }
  result:= make([]User, 0)
  err := conn.GetList(&result, query)
  ```

  ```
  result:= make([]User, 0)
  err := conn.GetList(&result, "id=?", 1002)
  ```

- func (conn *Conn) GetOrderedList(result interface{}, order string, query interface{}, args ...interface{})

  GetOrderedList根据条件查询排序记录。

  ```
  query := User{
      Name: "test",
  }
  result:= make([]User, 0)
  err := conn.GetOrderedList(&result, "id desc", query)
  ```

- func (conn *Conn) GetPaginationList(result interface{}, order string, page, pageSize int, query interface{}, args ...interface{}) error

  GetPaginationList根据条件分页查询。

  ```
  query := User{
      Name: "test",
  }
  result:= make([]User, 0)
  err := conn.GetPaginationList(&users, "id desc", 1, 2, "update_time > ? and update_time < ?", startTime, endTime)
  ```

- func (conn *Conn) ExecSQL(result interface{}, sql string, args ...interface{})

  ExecSQL执行原生sql。

  ```
  result:= make([]User, 0)
  err := conn.ExecSQL(&result, "select * from users")
  ```

- 待续......



ps：大家有其它需求或发现bug可以提issue哈，谢谢。
