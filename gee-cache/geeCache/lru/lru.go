package lru

import (
	"container/list"
)

type entry struct {
	key   string
	value Value
}

// Value Len()返回所占用的内存大小
type Value interface {
	Len() int
}

// Cache 是LRU缓存，并发不安全
type Cache struct {
	// 允许使用的最大内存
	maxBytes int64
	// 当前已经使用的内存
	nbytes int64
	// 双向链表list.List
	ll *list.List
	// 字典map
	mp map[string]*list.Element
	// 记录被移除时的回调函数，可以为nil
	OnEvicted func(key string, value Value)
}

func NewCache(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		mp:        make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.mp[key]; ok {
		// 如果键对应的链表节点存在，则将对应节点移动到队首
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, ok
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	// 取队尾节点，即最近最少访问的节点
	if ele := c.ll.Back(); ele != nil {
		kv := ele.Value.(*entry)
		// 从链表中删除该节点
		c.ll.Remove(ele)
		// 从字典cache中删除节点映射关系
		delete(c.mp, kv.key)
		// 更新当前所用内存
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 若回调函数不为nil，调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 同时实现新增和修改的功能
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.mp[key]; ok {
		// 若key已存在，修改
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 若key不存在，新增
		ele := c.ll.PushFront(&entry{key, value})
		c.mp[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
