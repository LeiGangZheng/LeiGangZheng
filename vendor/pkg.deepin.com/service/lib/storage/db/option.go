package db

// Option ...
type Option func(*Conn) error

// WithDialect 设置数据库类型
func WithDialect(dialect string) Option {
	return func(conn *Conn) error {
		conn.conf.Dialect = dialect
		return nil
	}
}

// WithDSN 设置数据库地址
func WithDSN(dsn string) Option {
	return func(conn *Conn) error {
		conn.conf.DSN = dsn
		return nil
	}
}

// WithMaxOpenConns 设置最大连接数
func WithMaxOpenConns(maxOpenConns int) Option {
	return func(conn *Conn) error {
		conn.conf.MaxOpenConns = maxOpenConns
		return nil
	}
}

// WithMaxIdleConns 设置最大闲置连接数
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(conn *Conn) error {
		conn.conf.MaxIdleConns = maxIdleConns
		return nil
	}
}

// WithConnMaxLifetime 设置最大回收时间
func WithConnMaxLifetime(connMaxLifetime int) Option {
	return func(conn *Conn) error {
		conn.conf.ConnMaxLifetime = connMaxLifetime
		return nil
	}
}

// WithLogMode 是否开启debug日志
func WithLogMode(logMode bool) Option {
	return func(conn *Conn) error {
		conn.conf.LogMode = logMode
		return nil
	}
}

// WithLogSQL 设置日志中是否包含sql字段
func WithLogSQL(logSQL bool) Option {
	return func(conn *Conn) error {
		conn.conf.LogSQL = logSQL
		return nil
	}
}

// WithSlowThreshold 设置慢日志阈值
func WithSlowThreshold(slowThreshold int) Option {
	return func(conn *Conn) error {
		conn.conf.SlowThreshold = slowThreshold
		return nil
	}
}

// WithSlaveFlag 设置DB角色是否为从DB
func WithSlaveFlag(flag bool) Option {
	return func(conn *Conn) error {
		conn.conf.SlaveFlag = flag
		return nil
	}
}

// DefaultConnConf 返回默认配置
func DefaultConnConf() *ConnConf {
	return &ConnConf{
		DSN:             "",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 300,
		LogMode:         false,
		LogSQL:          true,
		SlowThreshold:   500,
		SlaveFlag:       false,
	}
}
