package db

import (
	"github.com/jinzhu/gorm"
)

// GetOne 根据条件查询第一条记录，并返回是否存在
func (conn *Conn) GetOne(result interface{}, query interface{}, args ...interface{}) (isExist bool, err error) {
	err = conn.Where(query, args...).First(result).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// GetList 根据条件查询记录
func (conn *Conn) GetList(result interface{}, query interface{}, args ...interface{}) error {
	return conn.getListCore(result, "", 0, 0, query, args)
}

// GetOrderedList 根据条件查询排序记录
func (conn *Conn) GetOrderedList(result interface{}, order string, query interface{}, args ...interface{}) error {
	return conn.getListCore(result, order, 0, 0, query, args)
}

// GetFirstNRecords 根据条件查询前n条记录
func (conn *Conn) GetFirstNRecords(result interface{}, order string, limit int, query interface{}, args ...interface{}) error {
	return conn.getListCore(result, order, limit, 0, query, args)
}

// GetPageRangeList ...
func (conn *Conn) GetPageRangeList(result interface{}, order string, limit, offset int, query interface{}, args ...interface{}) error {
	return conn.getListCore(result, order, limit, offset, query, args)
}

// GetPaginationList 根据条件查询分页记录
func (conn *Conn) GetPaginationList(result interface{}, order string, page, pageSize int, query interface{}, args ...interface{}) error {
	return conn.getListCore(result, order, pageSize, (page-1)*pageSize, query, args)
}

func (conn *Conn) getListCore(result interface{}, order string, limit, offset int, query interface{}, args []interface{}) error {
	db := conn.Where(query, args...)
	if len(order) != 0 {
		db = db.Order(order)
	}

	if offset > 0 {
		db = db.Offset(offset)
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	if err := db.Find(result).Error; err != nil {
		return err
	}

	return nil
}

// ExecSQL 执行sql
func (conn *Conn) ExecSQL(result interface{}, sql string, args ...interface{}) error {
	err := conn.Raw(sql, args...).Scan(result).Error
	if err != nil {
		return err
	}

	return nil
}

// AutoPreload 设置开启自动预加载
func (conn *Conn) AutoPreload(auto bool) *Conn {
	return conn.Set("gorm:auto_preload", auto)
}

// ForUpdate 添加数据库排他锁
func (conn *Conn) ForUpdate() *Conn {
	return conn.Set("gorm:query_option", "FOR UPDATE")

}
