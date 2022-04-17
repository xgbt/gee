package consistentHash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map contains all hashed keys
type Map struct {
	hash     Hash           // 哈希函数
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环结构, sorted
	mp       map[int]string // 虚拟节点与真实节点的映射表，键是节点的哈希值，值是节点的名称
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		mp:       make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 对每个真实节点，创建replicas个虚拟节点
		for i := 0; i < m.replicas; i++ {
			// 通过编号i的方式区分不同的虚拟节点，然后计算虚拟节点哈希值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 把虚拟节点哈希值添加到环上
			m.keys = append(m.keys, hash)
			// 增加虚拟节点和真实节点的映射关系
			m.mp[hash] = key
		}
	}

	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}
	// 计算key的哈希值
	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个匹配的虚拟节点的下标 idx
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// m.keys是一个环形结构，所以使用取余数的方式
	return m.mp[m.keys[idx%len(m.keys)]]
}
