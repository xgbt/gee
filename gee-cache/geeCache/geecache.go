package geeCache

import (
	"fmt"
	pb "geeCache/geecachepb"
	"geeCache/singleflight"
	"log"
	"sync"
)

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Group 可以看成一个缓存的命名空间
type Group struct {
	name      string
	getter    Getter // 缓存未命中时获取源数据的回调(callback)
	mainCache cache  // 并发缓存
	peers     PeerPicker
	loader    *singleflight.Group // 加上 singleflight.Group，确保每个key只被请求一次
}

func NewGroup(name string, getter Getter, cacheBytes int64) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// 流程（1）：从 mainCache 中查找缓存，如果存在则返回缓存值。
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	// 流程（3）：缓存不存在，则调用 load 方法
	return g.load(key)
}

// RegisterPeers 将实现了PeerPicker接口的HTTPPool注入到Group中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	// 将实现了PeerPicker接口的HTTPPool注入到Group中
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 每个key只会被请求一次
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			// 使用PickPeer() 选择节点，若为ok则说明选择的节点为远程节点
			if peer, ok := g.peers.PickPeer(key); ok {
				// 调用getFromPeer获取缓存值
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		// !ok,说明选择远程节点失败或者选择的是本地节点
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}
	return
}

// getFromPeer 使用实现了PeerGetter接口的httpGetter访问远程节点，获取缓存值
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 调用用户回调函数 g.getter.Get() 获取源数据
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	// 并且将源数据添加到缓存 mainCache 中
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
