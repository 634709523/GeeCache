package lru

import (
	"container/list"
)

// 缓存是一个LRU缓存。它是不安全的并发访问。
type Cache struct {
	maxBytes  int64                         // 允许使用的最大内存
	nbytes    int64                         // 当前已使用的内存
	ll        *list.List                    // Go语言标准库实现的双向链表
	cache     map[string]*list.Element      // key是键,v是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数
}

// 双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

// 用于返回值所占用的内存大小
type Value interface {
	Len() int
}

// 实例化Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 查找功能
func (c *Cache) Get(key string) (value Value, ok bool) {
	// 从字典中找到对应的双向链表的节点
	if ele, ok := c.cache[key]; ok {
		// 如果存在，则将对应节点移动到队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 缓存淘汰:移除最近最少访问的节点(队首)
func (c *Cache) RemoveOldest() {
	// 取到队首节点 从链表中删除
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// 从字典中c.cache删除该节点的映射关系
		delete(c.cache, kv.key)
		// 更新当前所用的内存
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 如果缓存中已经有这个key,则更新对应节点的值，并将该节点移动到队尾
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		return
	}
	// 不存在新增
	ele := c.ll.PushFront(&entry{key: key, value: value})
	c.cache[key] = ele
	c.nbytes += int64(len(key)) + int64(value.Len())
	// 如果超过了设定的最大值，则移除最少访问节点
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 缓存双向链表的长度
func (c *Cache) Len() int {
	return c.ll.Len()
}
