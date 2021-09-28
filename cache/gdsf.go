package cache

import (
	"container/heap"
	"fmt"
	"log"
	"sync"
	"time"
)

type gdsfNode struct {
	key      string
	value    []byte
	count    int
	index    int
	priority float64
	size     int
	lastUsed time.Time
}

func (node *gdsfNode) DebugString() string {
	return fmt.Sprintf("%f; count: %d; size: %d; key: %s", node.priority, node.count, node.size, Shorten(node.key, 10))
}

// Priority Queue
type gdsfPQ []*gdsfNode

func (pq gdsfPQ) Len() int { return len(pq) }

func (pq gdsfPQ) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq gdsfPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *gdsfPQ) Push(x interface{}) {
	n := len(*pq)
	node := x.(*gdsfNode)
	node.index = n
	*pq = append(*pq, node)
}

func (pq *gdsfPQ) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil  // avoid memory leak
	node.index = -1 // for safety
	*pq = old[0 : n-1]
	return node
}

func (pq gdsfPQ) kth(k int) *gdsfNode {
	return pq[pq.Len()-k-1]
}

// Greedy Dual Size Frequency
type GDSFCache struct {
	maxSize      int
	maxEvictions int
	usedSize     int
	L            float64
	queue        gdsfPQ
	byKey        map[string]*gdsfNode

	mutex *sync.RWMutex
}

func NewGDSFCache(size int, maxEvictions int) *GDSFCache {
	return &GDSFCache{
		size,
		maxEvictions,
		0,
		0,
		gdsfPQ{},
		map[string]*gdsfNode{},
		new(sync.RWMutex),
	}
}

// evict element
func (cache *GDSFCache) aquisitionCost(size int) float64 {
	return float64(size)
}

func (cache *GDSFCache) cost(count int, size int) float64 {
	return cache.L + float64(count)*cache.aquisitionCost(size)/float64(size)
}

// evict element
func (cache *GDSFCache) evict(k int) error {
	for i := k; i > 0 && cache.queue.Len() > 0; i-- {
		node := heap.Pop(&cache.queue).(*gdsfNode)
		delete(cache.byKey, node.key)

		log.Printf("Evicting key %q; size %d; count: %d;duration: %q",
			node.key, node.size, node.count, time.Since(node.lastUsed))

		cache.usedSize -= node.size

		// replace L with the highest priority max_j { node_j.priority}
		if i == 1 {
			cache.L = node.priority
		}
	}
	return nil
}

// put blob into cache
func (cache *GDSFCache) Put(key string, value []byte) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	size := len(value)

	// check if we have enough space
	if size+cache.usedSize > cache.maxSize {
		evictSize := 0
		var itemsToEvict int
		for itemsToEvict = 1; itemsToEvict <= cache.maxEvictions && itemsToEvict < cache.queue.Len(); itemsToEvict++ {
			evictSize += cache.queue.kth(itemsToEvict - 1).size

			if size+cache.usedSize-evictSize > cache.maxSize {
				cache.evict(itemsToEvict)
				break
			}
		}
		if itemsToEvict > cache.maxEvictions {
			return fmt.Errorf("object is too big; can't free up enough memory with max evictions"+
				"size: %d; cache.maxEvictions: %d; maxEvictSize: %d; cacheSize: %d",
				size, cache.maxEvictions, evictSize, cache.maxSize)
		}
	}
	log.Printf("Admitted key %q; size %d", key, size)

	count := 1
	node := gdsfNode{key, value, count, 0, cache.cost(count, size), size, time.Now().UTC()}
	cache.byKey[key] = &node
	heap.Push(&cache.queue, &node)
	cache.usedSize += size
	return nil
}

// returns value for a given key
func (cache *GDSFCache) Get(key string) ([]byte, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if node, ok := cache.byKey[key]; ok {
		node.lastUsed = time.Now().UTC()
		node.count++
		node.priority = cache.cost(node.count, node.size)
		heap.Fix(&cache.queue, node.index)

		log.Printf("Get key %q; count %d; h: %f", key, node.count, node.priority)

		return node.value, true
	}
	return nil, false
}

// returns true if key is in cache; doesn't count as hit
func (cache *GDSFCache) Has(key string) bool {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	if _, ok := cache.byKey[key]; ok {
		log.Printf("Has key %q", key)
		return true
	}
	return false
}

func (cache *GDSFCache) Size() int {
	return cache.usedSize
}

func (cache *GDSFCache) DebugLog(k int) {
	for i := 0; i < k; i++ {
		log.Println(cache.queue.kth(i).DebugString())
	}
}
