package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"pkg.deepin.com/service/lib/log"

	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	_ "pkg.deepin.com/golang/lib/gorm"
)

var defaultConn *Conn

// Conn 单个DB连接
type Conn struct {
	*gorm.DB
	conf *ConnConf
}

// NewConn 新建一个DB连接
func NewConn(options ...Option) (conn *Conn, err error) {
	conn = &Conn{conf: DefaultConnConf()}

	for _, opt := range options {
		if err = opt(conn); err != nil {
			return
		}
	}

	return NewConn1(conn.conf)
}

// NewConn1 ...
func NewConn1(conf *ConnConf) (conn *Conn, err error) {
	conn = &Conn{conf: conf}

	if conf.DSN == "" {
		err = ERR_DB_DSN_EMPTY
		return
	}
	if conf.Dialect == "" {
		conf.Dialect = DEFAULT_DB_DIALECT
	}

	conn.DB, err = gorm.Open(conf.Dialect, conf.DSN)
	if err != nil {
		return
	}

	// 连接池设置
	if conf.MaxOpenConns > 0 {
		conn.DB.DB().SetMaxOpenConns(conf.MaxOpenConns)
	}

	if conf.MaxIdleConns > 0 {
		conn.DB.DB().SetMaxIdleConns(conf.MaxIdleConns)
	}

	if conf.ConnMaxLifetime > 0 {
		conn.DB.DB().SetConnMaxLifetime(time.Second * time.Duration(conf.ConnMaxLifetime))
	}

	// 设置日志
	conn.DB.LogMode(conf.LogMode)

	// 设置日志输出
	if conf.LogMode {
		if err := replaceCallback(conn); err != nil {
			return conn, err
		}
	}

	return
}

// InitConn 初始化一个DB连接
func InitConn(conf *ConnConf) {
	var err error
	defaultConn, err = NewConn1(conf)
	if err != nil {
		panic(err)
	}

	d, err := ParseDSN(defaultConn.conf)
	if err != nil {
		log.Error("parse dsn fail", "err", err.Error())
		return
	}

	log.Info("init DB success", "addr", d.Addr, "dbName", d.DBName, "user", d.User)
}

// Default 获取defaultConn实例
func Default() *Conn {
	return defaultConn
}

// replaceCallback 扩展原始callback打印日志
func replaceCallback(conn *Conn) (err error) {
	d, err := ParseDSN(conn.conf)
	if err != nil {
		log.Error("parse dsn fail", "err", err.Error())
		return
	}
	label := fmt.Sprintf("%s_%s", d.DBName, d.Addr)

	replace := func(processor func() *gorm.CallbackProcessor, callbackName string, wrapper func(func(*gorm.Scope), string) func(*gorm.Scope)) {
		old := processor().Get(callbackName)
		processor().Replace(callbackName, wrapper(old, label))
	}

	// TODO: prometheus metic
	invoke := func(op string) func(callback func(scope *gorm.Scope), label string) func(scope *gorm.Scope) {
		return func(callback func(scope *gorm.Scope), label string) func(scope *gorm.Scope) {
			return func(scope *gorm.Scope) {
				fn := func() {
					// 从数据库禁止create、update、delete操作
					if conn.conf.SlaveFlag && (op == "create" || op == "update" || op == "delete") {
						log.Error("error db log", "db_caller", new(Utils).FileWithLineNum(), "sql", new(Utils).HandleScopeSQL(conn.conf.LogSQL, conn.conf.Dialect, scope.SQL, scope.SQLVars), "tableName", scope.TableName(), "dbName", d.DBName, "addr", d.Addr, "rowsAffected", scope.DB().RowsAffected, "err", ERR_DB_SLAVE_FORBID_UPDATE.Error())
						return
					}

					beg := time.Now()
					callback(scope)
					cost := time.Since(beg)
					// 错误日志
					if scope.HasError() {
						if scope.DB().Error != gorm.ErrRecordNotFound {
							log.Error("error db log", "db_caller", new(Utils).FileWithLineNum(), "sql", new(Utils).HandleScopeSQL(conn.conf.LogSQL, conn.conf.Dialect, scope.SQL, scope.SQLVars), "cost", cost.String(), "tableName", scope.TableName(), "dbName", d.DBName, "addr", d.Addr, "rowsAffected", scope.DB().RowsAffected, "err", scope.DB().Error.Error())
							return
						}

						log.Warn("warn db log", "db_caller", new(Utils).FileWithLineNum(), "sql", new(Utils).HandleScopeSQL(conn.conf.LogSQL, conn.conf.Dialect, scope.SQL, scope.SQLVars), "cost", cost.String(), "tableName", scope.TableName(), "dbName", d.DBName, "addr", d.Addr, "rowsAffected", scope.DB().RowsAffected, "err", scope.DB().Error.Error())
						return
					}
					// 慢日志
					if conn.conf.SlowThreshold > 0 && cost > time.Duration(conn.conf.SlowThreshold)*time.Millisecond {
						log.Warn("slow db log", "db_caller", new(Utils).FileWithLineNum(), "sql", new(Utils).HandleScopeSQL(conn.conf.LogSQL, conn.conf.Dialect, scope.SQL, scope.SQLVars), "cost", cost.String(), "tableName", scope.TableName(), "dbName", d.DBName, "addr", d.Addr, "rowsAffected", scope.DB().RowsAffected)
						return
					}
					// debug日志
					if conn.conf.LogMode {
						log.Debug("debug db log", "db_caller", new(Utils).FileWithLineNum(), "sql", new(Utils).HandleScopeSQL(conn.conf.LogSQL, conn.conf.Dialect, scope.SQL, scope.SQLVars), "cost", cost.String(), "tableName", scope.TableName(), "dbName", d.DBName, "addr", d.Addr, "rowsAffected", scope.DB().RowsAffected)
						return
					}
				}

				fn()
			}
		}
	}

	replace(conn.DB.Callback().Delete, "gorm:delete", invoke("delete"))
	replace(conn.DB.Callback().Update, "gorm:update", invoke("update"))
	replace(conn.DB.Callback().Create, "gorm:create", invoke("create"))
	replace(conn.DB.Callback().Query, "gorm:query", invoke("query"))
	replace(conn.DB.Callback().RowQuery, "gorm:row_query", invoke("row_query"))

	return
}

//===================================================== gorm =====================================================
// Where ...
func (conn *Conn) Where(query interface{}, args ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Where(query, args...),
		conn.conf,
	}
}

// Or ...
func (conn *Conn) Or(query interface{}, args ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Or(query, args...),
		conn.conf,
	}
}

// Not ...
func (conn *Conn) Not(query interface{}, args ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Not(query, args...),
		conn.conf,
	}
}

// Limit ...
func (conn *Conn) Limit(limit interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Limit(limit),
		conn.conf,
	}
}

// Offset ...
func (conn *Conn) Offset(offset interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Offset(offset),
		conn.conf,
	}
}

// Order ...
func (conn *Conn) Order(value interface{}, reorder ...bool) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Order(value, reorder...),
		conn.conf,
	}
}

// Select ...
func (conn *Conn) Select(query interface{}, args ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Select(query, args...),
		conn.conf,
	}
}

// Omit ...
func (conn *Conn) Omit(columns ...string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Omit(columns...),
		conn.conf,
	}
}

// Group ...
func (conn *Conn) Group(query string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Group(query),
		conn.conf,
	}
}

// Having ...
func (conn *Conn) Having(query interface{}, values ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Having(query, values...),
		conn.conf,
	}
}

// Joins ...
func (conn *Conn) Joins(query string, args ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Joins(query, args...),
		conn.conf,
	}
}

// Scopes ...
func (conn *Conn) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	conn.DB = conn.DB.Scopes(funcs...)
	return conn
}

// Unscoped ...
func (conn *Conn) Unscoped() *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Unscoped(),
		conn.conf,
	}
}

// Attrs ...
func (conn *Conn) Attrs(attrs ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Attrs(attrs...),
		conn.conf,
	}
}

// Assign ...
func (conn *Conn) Assign(attrs ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Assign(attrs...),
		conn.conf,
	}
}

// First ...
func (conn *Conn) First(out interface{}, where ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.First(out, where...),
		conn.conf,
	}
}

// Take ...
func (conn *Conn) Take(out interface{}, where ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Take(out, where...),
		conn.conf,
	}
}

// Last ...
func (conn *Conn) Last(out interface{}, where ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Last(out, where...),
		conn.conf,
	}
}

// Find ...
func (conn *Conn) Find(out interface{}, where ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Find(out, where...),
		conn.conf,
	}
}

// Preloads ...
func (conn *Conn) Preloads(out interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Preloads(out),
		conn.conf,
	}
}

// Scan ...
func (conn *Conn) Scan(dest interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Scan(dest),
		conn.conf,
	}
}

// Pluck ...
func (conn *Conn) Pluck(column string, value interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Pluck(column, value),
		conn.conf,
	}
}

// Count ...
func (conn *Conn) Count(value interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Count(value),
		conn.conf,
	}
}

// Related ...
func (conn *Conn) Related(value interface{}, foreignKeys ...string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Related(value, foreignKeys...),
		conn.conf,
	}
}

// FirstOrInit ...
func (conn *Conn) FirstOrInit(out interface{}, where ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.FirstOrInit(out, where...),
		conn.conf,
	}
}

// FirstOrCreate ...
func (conn *Conn) FirstOrCreate(out interface{}, where ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.FirstOrCreate(out, where...),
		conn.conf,
	}
}

// Update ...
func (conn *Conn) Update(attrs ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Update(attrs...),
		conn.conf,
	}
}

// Updates ...
func (conn *Conn) Updates(values interface{}, ignoreProtectedAttrs ...bool) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Updates(values, ignoreProtectedAttrs...),
		conn.conf,
	}
}

// UpdateColumn ...
func (conn *Conn) UpdateColumn(attrs ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.UpdateColumn(attrs...),
		conn.conf,
	}
}

// UpdateColumns ...
func (conn *Conn) UpdateColumns(values interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.UpdateColumns(values),
		conn.conf,
	}
}

// Save ...
func (conn *Conn) Save(value interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Save(value),
		conn.conf,
	}

}

// Create ...
func (conn *Conn) Create(value interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Create(value),
		conn.conf,
	}
}

// Delete ...
func (conn *Conn) Delete(value interface{}, where ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Delete(value, where...),
		conn.conf,
	}
}

// Raw ...
func (conn *Conn) Raw(sql string, values ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Raw(sql, values...),
		conn.conf,
	}
}

// Exec ...
func (conn *Conn) Exec(sql string, values ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Exec(sql, values...),
		conn.conf,
	}
}

// Model ...
func (conn *Conn) Model(value interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Model(value),
		conn.conf,
	}
}

// Table ...
func (conn *Conn) Table(name string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Table(name),
		conn.conf,
	}
}

// Begin ...
func (conn *Conn) Begin() *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Begin(),
		conn.conf,
	}
}

// BeginTx ...
func (conn *Conn) BeginTx(ctx context.Context, opts *sql.TxOptions) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.BeginTx(ctx, opts),
		conn.conf,
	}
}

// Commit ...
func (conn *Conn) Commit() *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	conn.DB = conn.DB.Commit()
	return conn
}

// Rollback ...
func (conn *Conn) Rollback() *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	conn.DB = conn.DB.Rollback()
	return conn
}

// RollbackUnlessCommitted ...
func (conn *Conn) RollbackUnlessCommitted() *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	conn.DB = conn.DB.RollbackUnlessCommitted()
	return conn
}

// CreateTable ...
func (conn *Conn) CreateTable(models ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.CreateTable(models...),
		conn.conf,
	}
}

// DropTable ...
func (conn *Conn) DropTable(values ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.DropTable(values...),
		conn.conf,
	}
}

// DropTableIfExists ...
func (conn *Conn) DropTableIfExists(values ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.DropTableIfExists(values...),
		conn.conf,
	}
}

// AutoMigrate ...
func (conn *Conn) AutoMigrate(values ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.AutoMigrate(values...),
		conn.conf,
	}

}

// ModifyColumn ...
func (conn *Conn) ModifyColumn(column string, typ string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.ModifyColumn(column, typ),
		conn.conf,
	}
}

// DropColumn ...
func (conn *Conn) DropColumn(column string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.DropColumn(column),
		conn.conf,
	}
}

// AddIndex ...
func (conn *Conn) AddIndex(indexName string, columns ...string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.AddIndex(indexName, columns...),
		conn.conf,
	}
}

// AddUniqueIndex ...
func (conn *Conn) AddUniqueIndex(indexName string, columns ...string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.AddUniqueIndex(indexName, columns...),
		conn.conf,
	}
}

// RemoveIndex ...
func (conn *Conn) RemoveIndex(indexName string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.RemoveIndex(indexName),
		conn.conf,
	}
}

// AddForeignKey ...
func (conn *Conn) AddForeignKey(field string, dest string, onDelete string, onUpdate string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.AddForeignKey(field, dest, onDelete, onUpdate),
		conn.conf,
	}
}

// RemoveForeignKey ...
func (conn *Conn) RemoveForeignKey(field string, dest string) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.RemoveForeignKey(field, dest),
		conn.conf,
	}
}

// Preload ...
func (conn *Conn) Preload(column string, conditions ...interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Preload(column, conditions...),
		conn.conf,
	}
}

// Set ...
func (conn *Conn) Set(name string, value interface{}) *Conn {
	if conn.DB == nil || conn.DB.Error == ERR_DB_CONN_NIL {
		conn.DB = &gorm.DB{Error: ERR_DB_CONN_NIL}
		return conn
	}
	return &Conn{
		conn.DB.Set(name, value),
		conn.conf,
	}
}

// GetErrors ...
func (conn *Conn) GetErrors() []error {
	if conn.DB == nil {
		return []error{ERR_DB_CONN_NIL}
	}
	return conn.DB.GetErrors()
}
