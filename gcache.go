package gcache

import (
	"fmt"
	pb "gcache/gcachepb"
	"gcache/singleflight"
	"log"
	"sync"
)

// Group 是一个缓存命名空间及其关联的数据加载器
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	// 使用 singleflight.Group 确保每个键只被获取一次
	loader *singleflight.Group
}

// Getter 是一个接口，用于从数据源加载键对应的值
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 是一个实现了 Getter 接口的函数类型
type GetterFunc func(key string) ([]byte, error)

// Get 实现了 Getter 接口的方法
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建一个新的 Group 实例
// 参数:
// - name: 组的名称
// - cacheBytes: 缓存的最大字节数
// - getter: 数据加载器
// 返回:
// - *Group: 新创建的 Group 实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
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

// GetGroup 返回之前通过 NewGroup 创建的指定名称的 Group 实例
// 参数:
// - name: 组的名称
// 返回:
// - *Group: 指定名称的 Group 实例，如果不存在则返回 nil
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 从缓存中获取指定键的值
// 参数:
// - key: 键
// 返回:
// - ByteView: 键对应的值
// - error: 如果发生错误则返回错误
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GCache] hit")
		return v, nil
	}

	return g.load(key)
}

// RegisterPeers 注册一个 PeerPicker 以选择远程节点
// 参数:
// - peers: PeerPicker 实例
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// load 加载指定键的值，确保每个键只被加载一次
// 参数:
// - key: 键
// 返回:
// - ByteView: 键对应的值
// - error: 如果发生错误则返回错误
func (g *Group) load(key string) (value ByteView, err error) {
	// 每个键只被获取一次（无论是本地还是远程）
	// 不管有多少并发调用者
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// populateCache 将值填充到缓存中
// 参数:
// - key: 键
// - value: 值
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// getLocally 从本地数据源获取键对应的值
// 参数:
// - key: 键
// 返回:
// - ByteView: 键对应的值
// - error: 如果发生错误则返回错误
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// getFromPeer 从远程节点获取键对应的值
// 参数:
// - peer: 远程节点
// - key: 键
// 返回:
// - ByteView: 键对应的值
// - error: 如果发生错误则返回错误
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