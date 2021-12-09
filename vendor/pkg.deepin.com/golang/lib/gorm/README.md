这个是对gorm 1.x版本的一个拓展，主要是用于支持批量插入的操作。

需要注意的是，由于是批量操作，对于自增序列的返回值获取主要分为下面的情况：
1. 支持insert output/returning 的数据库
2. 不支持insert output的数据库

不支持insert output模式，则无法获取到返回值,对于支持insert output模式的数据库可以通过下面的方式进行获取

```
// 默认情况下，insert output模式是关闭的，只有数据库支持insert output模式时，该功能才有用
instance := db.New()
instance.InstantSet("returning", true)     //开启返回模式
instance.InstantSet("returning_all", true) //开启全返回模式，默认情况下，返回primary key和自动增长列
newinstance := instance.Create()                          //新增记录(单条记录走gorm默认实现，slice或array走extend实现) create会创建新的db instance
result, has := newinstance.Get("returning_value") //result为数据库返回记录结构化
```