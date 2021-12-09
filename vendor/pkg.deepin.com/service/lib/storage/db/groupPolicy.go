package db

import (
	"math/rand"
	"sync"
	"time"
)

// GroupPolicy 选择slave接口
type GroupPolicy interface {
	Slave(*Group) *Conn
}

type GroupPolicyHandler func(*Group) *Conn

// Slave 实现了slave的选择
func (h GroupPolicyHandler) Slave(g *Group) *Conn {
	return h(g)
}

// RandomPolicy 实现了随机选择slave
func RandomPolicy() GroupPolicyHandler {
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))
	return func(g *Group) *Conn {
		idx := r.Intn(len(g.Slaves()))
		return getAvailableConn(g, idx, 0)
	}
}

// WeightRandomPolicy 实现了根据权重随机选择slave
func WeightRandomPolicy(weights []int) GroupPolicyHandler {
	var rands = make([]int, 0, len(weights))
	for i := 0; i < len(weights); i++ {
		for n := 0; n < weights[i]; n++ {
			rands = append(rands, i)
		}
	}
	var r = rand.New(rand.NewSource(time.Now().UnixNano()))

	return func(g *Group) *Conn {
		var slaves = g.Slaves()
		idx := rands[r.Intn(len(rands))]
		if idx >= len(slaves) {
			idx = len(slaves) - 1
		}
		return getAvailableConn(g, idx, 0)
	}
}

// RoundRobinPolicy 实现了轮询选择slave
func RoundRobinPolicy() GroupPolicyHandler {
	var pos = -1
	var lock sync.Mutex
	return func(g *Group) *Conn {
		var slaves = g.Slaves()

		lock.Lock()
		defer lock.Unlock()
		pos++
		if pos >= len(slaves) {
			pos = 0
		}

		return getAvailableConn(g, pos, 0)
	}
}

// getAvailableConn 从库进行故障检测，返回的一直是可用的节点
func getAvailableConn(g *Group, r int, dep int) *Conn {
	count := len(g.Slaves())

	// 递归深度跟节点数相同，则所有从节点都挂，无可用节点了，用master节点
	if dep == count {
		return g.Conn
	}

	db := g.Slaves()[r]
	err := db.DB.DB().Ping()
	if err != nil {
		if r < count-1 {
			return getAvailableConn(g, r+1, dep+1)
		}

		if r == count-1 {
			return getAvailableConn(g, 0, dep+1)
		}
	}

	return db
}
