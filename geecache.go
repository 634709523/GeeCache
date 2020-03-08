package geecache

import (
	"fmt"
	"log"
	"sync"
)

// Getter loads data for a key
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 实现Getter函数
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}


// 缓存的命名空间
type Group struct {
	name      string // 唯一名称
	getter    Getter // 缓存未命中时获取源数据的回调
	mainCache cache  // 一开始实现的并发缓存
}

func (g *Group)Get(key string)(ByteView,error){
	if key == ""{
		return ByteView{},fmt.Errorf("key is required")
	}
	// 从 mainCache 中查找缓存，如果存在则返回缓存值
	if v, ok := g.mainCache.Get(key);ok{
		log.Println("[GeeCache] hit")
		return v,nil
	}
	// 缓存不存在，则调用 load 方法
	return g.load(key)
}

func (g *Group)load (key string)(value ByteView,err error){
	return g.getLocallly(key)
}

func (g *Group)getLocallly(key string)(ByteView,error){
	bytes, err := g.getter.Get(key)
	if err != nil{
		return ByteView{},err
	}
	value := ByteView{b:cloneBytes(bytes)}
	g.populateCache(key,value)
	return value,nil
}

// populateCache 源数据添加到缓存 mainCache 中
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}










var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 初始化group，将 group 存储在全局变量 groups 中
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup 获取特定名称的Group
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.Unlock()
	return g
}
