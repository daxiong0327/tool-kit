package singleflight

import (
	"golang.org/x/sync/singleflight"
)

// DefaultGroup 是默认的 Group 实例
var DefaultGroup = NewGroup()

// Group 代表一个 singleflight 实例
type Group struct {
	sf singleflight.Group
}

// NewGroup 创建一个新的 Group 实例
func NewGroup() *Group {
	return &Group{
		sf: singleflight.Group{},
	}
}

// Do 执行函数并返回结果。如果有相同的 key 正在执行，会等待已有的执行完成并返回其结果
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error, bool) {
	return g.sf.Do(key, fn)
}

// DoChan 类似于 Do，但返回一个 channel 来接收结果
func (g *Group) DoChan(key string, fn func() (interface{}, error)) <-chan singleflight.Result {
	return g.sf.DoChan(key, fn)
}

// Forget 从 Group 中删除 key 对应的进行中的操作
func (g *Group) Forget(key string) {
	g.sf.Forget(key)
}
