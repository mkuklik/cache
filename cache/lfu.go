package cache

import (
	"container/heap"
	"log"
	"sync"
	"time"
)

type lfuNode struct {
	key      string
	value    []byte
	count    int
	index    int
	lastUsed time.Time
}

// Priority Queue
type PQ []*lfuNode

func (pq PQ) Len() int { return len(pq) }

func (pq PQ) Less(i, j int) bool {
	return pq[i].count < pq[j].count ||
		(pq[i].count == pq[j].count && pq[i].lastUsed.Before(pq[j].lastUsed))
}

func (pq PQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PQ) Push(x interface{}) {
	n := len(*pq)
	node := x.(*lfuNode)
	node.index = n
	*pq = append(*pq, node)
}

func (pq *PQ) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil  // avoid memory leak
	node.index = -1 // for safety
	*pq = old[0 : n-1]
	return node
}

// Least Frequently Used
type LFUCache struct {
	maxSize  int
	currSize int
	pq       PQ
	byKey    map[string]*lfuNode

	mutex *sync.RWMutex
}

func NewLFUCache(size int) *LFUCache {
	return &LFUCache{
		size,
		0,
		PQ{},
		map[string]*lfuNode{},
		new(sync.RWMutex),
	}
}

// evict element
func (cache *LFUCache) evict() error {
	if cache.pq.Len() > 0 {
		node := heap.Pop(&cache.pq).(*lfuNode)
		delete(cache.byKey, node.key)
		size := len(node.value)

		log.Printf("Evicting key %q; size %d; count: %d;duration: %q",
			node.key, size, node.count, time.Since(node.lastUsed))

		cache.currSize -= size
	}
	return nil
}

// put blob into cache
func (cache *LFUCache) Put(key string, value []byte) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// check if we have enough space
	size := len(value)
	if size+cache.currSize > cache.maxSize {
		cache.evict()
	}

	log.Printf("Admitted key %q; size %d", key, size)

	node := lfuNode{key, value, 1, 0, time.Now().UTC()}
	cache.byKey[key] = &node
	heap.Push(&cache.pq, &node)
	cache.currSize += size
	return nil
}

// returns value for a given key
func (cache *LFUCache) Get(key string) ([]byte, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if node, ok := cache.byKey[key]; ok {
		node.lastUsed = time.Now().UTC()
		node.count++
		heap.Fix(&cache.pq, node.index)

		log.Printf("Get key %q; count %d", key, node.count)

		return node.value, true
	}
	return nil, false
}

// returns true if key is in cache; doesn't count as hit
func (cache *LFUCache) Has(key string) bool {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	if _, ok := cache.byKey[key]; ok {
		log.Printf("Has key %q", key)
		return true
	}
	return false
}

func (cache *LFUCache) Size() int {
	return cache.currSize
}
