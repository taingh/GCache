package singleflight

import "sync"

// call 表示正在进行或已完成的 Do 调用
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 表示一类工作，并形成一个命名空间，在该命名空间内可以执行工作并抑制重复执行。
type Group struct {
	mu sync.Mutex       // 保护 m
	m  map[string]*call // 懒初始化
}

// Do 执行给定的函数并返回结果，确保对于给定的键同时只有一个执行。
// 如果有重复的调用，重复的调用者会等待原始调用完成并接收相同的结果。
// 参数:
//   key: 唯一标识任务的键
//   fn: 需要执行的函数，返回一个接口类型的值和错误
// 返回值:
//   val: 函数执行的结果
//   err: 函数执行过程中遇到的错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// 执行函数并获取结果
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}