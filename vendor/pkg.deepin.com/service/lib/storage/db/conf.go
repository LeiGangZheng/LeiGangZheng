package db

// Conf DBManager配置
type Conf map[string]GroupConf

// GroupConf DB主从配置
type GroupConf struct {
	Master MasterConf
	Slave  SlaveConf
}

// MasterConf Master配置
type MasterConf struct {
	Dialect         string // 类型
	DSN             string // 地址
	MaxOpenConns    int    // 最大连接数
	MaxIdleConns    int    // 最大闲置连接数
	ConnMaxLifetime int    // 最大空闲回收时间
	LogMode         bool   // 是否开启debug日志
	LogSQL          bool   // 是否显示sql中具体值
	SlowThreshold   int    // 慢日志阈值
}

// SlaveConf Slave配置
type SlaveConf struct {
	Dialect         string   // 类型
	DSN             []string // 地址
	MaxOpenConns    int      // 最大连接数
	MaxIdleConns    int      // 最大闲置连接数
	ConnMaxLifetime int      // 最大空闲回收时间
	LogMode         bool     // 是否开启debug日志
	LogSQL          bool     // 是否显示日志中的sql
	SlowThreshold   int      // 慢日志阈值
}

// ConnConf 连接配置
type ConnConf struct {
	Dialect         string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	LogMode         bool
	LogSQL          bool
	SlowThreshold   int
	SlaveFlag       bool
}
