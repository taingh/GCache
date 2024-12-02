package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// Map 包含所有已哈希的键及其对应的节点
// 实现了一致性哈希算法以将键分布到节点上
type Map struct {
	hash     Hash
	replicas int
	keys     []int // 必须保持排序以便进行二分查找
	hashMap  map[int]string
}

// New 创建一个 Map 实例
// 参数 replicas 指定每个键的副本数量
// 参数 fn 是自定义的哈希函数，如果为 nil 则使用 crc32.ChecksumIEEE
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 向哈希映射中添加一组键
// 为每个键创建多个副本以实现虚拟节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 获取与提供的键最接近的项
// 参数 key 是要查找的键
// 返回与键最接近的节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 对副本进行二分查找
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}