package db

import (
	"github.com/pkg/errors"
)

// Group DB组
type Group struct {
	*Conn
	slaves []*Conn
	policy GroupPolicy
}

// NewGroup 新建DB组
func NewGroup(groupConf *GroupConf, policies ...GroupPolicy) (group *Group, err error) {
	var dbConn *Conn
	group = &Group{}

	if len(policies) > 0 {
		group.policy = policies[0]
	} else {
		group.policy = RandomPolicy()
	}

	//master和slave配置
	masterConf := groupConf.Master
	slaveConf := groupConf.Slave

	// master连接
	if masterConf.DSN == "" {
		err = ERR_DB_MASTER_DSN_EMPTY
		return
	}
	dbConn, err = NewConn(
		WithDialect(masterConf.Dialect),
		WithDSN(masterConf.DSN),
		WithMaxOpenConns(masterConf.MaxOpenConns),
		WithMaxIdleConns(masterConf.MaxIdleConns),
		WithConnMaxLifetime(masterConf.ConnMaxLifetime),
		WithLogMode(masterConf.LogMode),
		WithLogSQL(masterConf.LogSQL),
		WithSlowThreshold(masterConf.SlowThreshold),
	)
	if err != nil {
		return
	}
	group.Conn = dbConn

	// slave连接池
	if slaveConf.DSN != nil {
		for i := 0; i < len(slaveConf.DSN); i++ {
			dbConn, err = NewConn(
				WithDialect(slaveConf.Dialect),
				WithDSN((slaveConf.DSN)[i]),
				WithMaxOpenConns(slaveConf.MaxOpenConns),
				WithMaxIdleConns(slaveConf.MaxIdleConns),
				WithConnMaxLifetime(slaveConf.ConnMaxLifetime),
				WithLogMode(slaveConf.LogMode),
				WithLogSQL(slaveConf.LogSQL),
				WithSlowThreshold(masterConf.SlowThreshold),
				WithSlaveFlag(true),
			)
			if err != nil {
				return
			}
			group.slaves = append(group.slaves, dbConn)
		}
	}
	return
}

// SetPolicy 设置获取slave的策略
func (group *Group) SetPolicy(policy GroupPolicy) {
	group.policy = policy
	return
}

// Master 获取Master连接
func (group *Group) Master() (conn *Conn) {
	return group.Conn
}

// Slave 根据策略获取某slave连接
func (group *Group) Slave() (conn *Conn) {
	switch len(group.slaves) {
	case 0:
		return group.Conn
	case 1:
		return group.slaves[0]
	}
	return group.policy.Slave(group)
}

// Slaves 获取所有slave连接
func (group *Group) Slaves() []*Conn {
	return group.slaves
}

// Close 关闭master和slave DB
func (group *Group) Close() (err error) {
	if err = group.Conn.DB.Close(); err != nil {
		return
	}
	for _, v := range group.slaves {
		if e := v.DB.Close(); e != nil {
			err = errors.WithStack(e)
		}
	}
	return
}

// Ping 验证连接是否正常
func (group *Group) Ping() (err error) {
	if err = group.Conn.DB.DB().Ping(); err != nil {
		return
	}
	for _, v := range group.slaves {
		if err = v.DB.DB().Ping(); err != nil {
			return
		}
	}
	return
}
