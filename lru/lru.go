package lru

import "container/list"

// Cache 是一个LRU缓存。它不支持并发访问。
type Cache struct {
	maxBytes int64
	nbytes   int64
	ll       *list.List
	cache    map[string]*list.Element
	// 当条目被清除时可选执行。
	OnEvicted func(key string, value Value)
}

// entry 是缓存中的条目结构。
type entry struct {
	key   string
	value Value
}

// Value 是一个接口，用于计算值占用的字节数。
type Value interface {
	Len() int
}

// New 是 Cache 的构造函数。
// 参数:
//   maxBytes: 缓存的最大字节数。
//   onEvicted: 当条目被清除时调用的回调函数。
// 返回:
//   *Cache: 新创建的缓存实例。
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add 向缓存中添加一个值。
// 参数:
//   key: 要添加的键。
//   value: 要添加的值。
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get 查找键的值。
// 参数:
//   key: 要查找的键。
// 返回:
//   value: 键对应的值。
//   ok: 是否找到键。
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 移除最老的项。
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len 返回缓存中的条目数量。
// 返回:
//   int: 缓存中的条目数量。
func (c *Cache) Len() int {
	return c.ll.Len()
}