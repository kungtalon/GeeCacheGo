package coordinator

import (
	"hash/crc32"
	"io/ioutil"
	"sort"
	"strconv"
	"sync"
)

type Hash func(data []byte) uint32

type HashRing struct {
	hash       Hash
	replicas   int            // number of virtual nodes replicas
	hashKeys   []int          // sorted list of hash keys
	hashMap    map[int]string // hashMap from virtual node key to real key
	privateKey string         // privateKey used to avoid hash flooding attack
	mu         sync.RWMutex
}

func New(replicas int, privateKey string, fn Hash) *HashRing {
	m := &HashRing{
		replicas:   replicas,
		hash:       fn,
		hashMap:    make(map[int]string),
		privateKey: privateKey,
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds servers to the consistent hashing ring
func (m *HashRing) Add(keys ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := m.generateHash(key, i)
			m.hashKeys = append(m.hashKeys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.hashKeys)
}

// Get returns the next closest hash key on the hash ring
func (m *HashRing) Get(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.hashKeys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.hashKeys), func(i int) bool {
		return m.hashKeys[i] >= hash
	})

	return m.hashMap[m.hashKeys[idx%len(m.hashKeys)]]
}

// Remove use to remove a key and its virtual keys on the ring and map
func (m *HashRing) Remove(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < m.replicas; i++ {
		hash := m.generateHash(key, i)
		idx := sort.SearchInts(m.hashKeys, hash)
		m.hashKeys = append(m.hashKeys[:idx], m.hashKeys[idx+1:]...)
		delete(m.hashMap, hash)
	}
}

func (m *HashRing) generateHash(key string, index int) int {
	return int(m.hash([]byte(strconv.Itoa(index) + m.privateKey + key)))
}

func readPrivateKey() (privateKey string) {
	content, err := ioutil.ReadFile("private.txt")
	if err != nil {
		return ""
	}

	privateKey = string(content)
	if len(privateKey) == 0 {
		return ""
	}
	return
}
