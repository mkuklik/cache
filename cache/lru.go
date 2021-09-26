package cache

import (
	"container/list"
	"log"
	"sync"
	"time"
)

type node struct {
	key      string
	value    []byte
	lastUsed time.Time
}

// Least Recently Used
type LRUCache struct {
	maxSize  int
	currSize int
	list     *list.List
	byKey    map[string]*list.Element

	mutex *sync.RWMutex
}

func NewLRUCache(size int) *LRUCache {
	return &LRUCache{
		size,
		0,
		list.New(),
		map[string]*list.Element{},
		new(sync.RWMutex),
	}
}

// evict element
func (cache *LRUCache) evict() error {
	if cache.list.Len() > 0 {
		elm := cache.list.Back()
		node := elm.Value.(node)
		size := len(node.value)

		log.Printf("Evicting key %q; size %d; duration: %q", node.key, size, time.Since(node.lastUsed))

		delete(cache.byKey, node.key)
		cache.list.Remove(elm)
		cache.currSize -= size
	}
	return nil
}

// put blob into cache
func (cache *LRUCache) Put(key string, value []byte) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// check if we have enough space
	size := len(value)
	if size+cache.currSize > cache.maxSize {
		cache.evict()
	}

	log.Printf("Admitted key %q; size %d", key, size)

	elm := cache.list.PushFront(node{key, value, time.Now().UTC()})
	cache.byKey[key] = elm
	cache.currSize += size
	return nil
}

func (cache *LRUCache) Get(key string) ([]byte, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if elm, ok := cache.byKey[key]; ok {
		cache.list.MoveToFront(elm)
		node := elm.Value.(node)
		node.lastUsed = time.Now().UTC()

		log.Printf("Get key %q", key)

		return node.value, true
	}
	return nil, false
}

// returns true if key is in cache; doesn't count as hit
func (cache *LRUCache) Has(key string) bool {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	if _, ok := cache.byKey[key]; ok {
		return true
	}
	return false
}

func (cache *LRUCache) Size() int {
	return cache.currSize
}

// linked list for quick access
// random access: map to nodes
