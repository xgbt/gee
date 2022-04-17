package singleflight

import "sync"

// 正在进行中、或者已经结束的请求
type call struct {
	wg  sync.WaitGroup // 避免重入
	val interface{}
	err error
}

// Group 主数据结构，管理不同key的请求（call）
type Group struct {
	mu sync.Mutex // 为保护mp不被并发读写而加上的锁
	mp map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.mp == nil {
		g.mp = make(map[string]*call)
	}
	if c, ok := g.mp[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // 如果请求正在进行中，则等待
		return c.val, c.err // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)   // 发起请求前加锁
	g.mp[key] = c // 添加到g.mp，表明key已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束

	g.mu.Lock()
	delete(g.mp, key) // 更新g.mp
	g.mu.Unlock()

	return c.val, c.err
}
