package db

import "errors"

// 初始化默认参数值
const (
	DEFAULT_DB_DIALECT string = "mysql" // 默认数据库类型为mysql
)

// 错误码
var (
	ERR_DB_MANAGER_NOT_FOUND   = errors.New("DB Manager is nil")
	ERR_DB_GROUP_NOT_FOUND     = errors.New("DB Group is nil")
	ERR_DB_DSN_EMPTY           = errors.New("DSN can not be empty")
	ERR_DB_MASTER_DSN_EMPTY    = errors.New("MasterDB DSN can not be empty")
	ERR_DB_CONN_NIL            = errors.New("DB Conn is nil")
	ERR_DSN_INVALID            = errors.New("DSN is invalid")
	ERR_DB_SLAVE_FORBID_UPDATE = errors.New("SlaveDB forbids to modify data")
)
