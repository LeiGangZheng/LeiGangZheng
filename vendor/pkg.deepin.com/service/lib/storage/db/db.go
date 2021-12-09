package db

import "github.com/pkg/errors"

var defaultManager *Manager

// Manager DB管理器
type Manager struct {
	groupMap map[string]*Group
}

// NewManager 新建一个DB管理器
func NewManager(config *Conf) (*Manager, error) {
	dbMgr := &Manager{
		groupMap: make(map[string]*Group),
	}

	if config == nil {
		return nil, errors.New("config can not be nil")
	}

	// 按Name索引每一个DB Group
	for k, v := range *config {
		var group *Group
		var err error
		if group, err = NewGroup(&v); err != nil {
			return nil, err
		}
		dbMgr.groupMap[k] = group
	}

	return dbMgr, nil
}

// InitManager 初始化defaultManager
func InitManager(config *Conf) {
	dbMgr, err := NewManager(config)
	if err != nil {
		panic(err)
	}
	defaultManager = dbMgr
}

// Group 根据name获取DB组
func (dbMgr *Manager) Group(name string) (group *Group) {
	return dbMgr.groupMap[name]
}

// Master 获取master连接
func (dbMgr *Manager) Master(name string) *Conn {
	if dbMgr.Group(name) == nil {
		return &Conn{}
	}

	return dbMgr.Group(name).Master()
}

// Slave 获取slave连接
func (dbMgr *Manager) Slave(name string) *Conn {
	if dbMgr.Group(name) == nil {
		return &Conn{}
	}
	return dbMgr.Group(name).Slave()
}

// DefaultManager 获取defaultMgr实例
func DefaultManager() *Manager {
	return defaultManager
}

// Master 获取defaultManager的master连接
func Master(name string) *Conn {
	if defaultManager.Group(name) == nil {
		return &Conn{}
	}

	return defaultManager.Group(name).Master()
}

// Slave 获取defaultManager的slave连接
func Slave(name string) *Conn {
	if defaultManager.Group(name) == nil {
		return &Conn{}
	}
	return defaultManager.Group(name).Slave()
}
