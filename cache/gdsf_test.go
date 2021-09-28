package cache

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestGDSFCache(t *testing.T) {

	key1 := "key1"
	val1 := make([]byte, 50)
	rand.Read(val1)

	key2 := "key2"
	val2 := make([]byte, 40)
	rand.Read(val2)

	key3 := "key3"
	val3 := make([]byte, 30)
	rand.Read(val3)

	t.Run("basic check", func(t *testing.T) {
		size := 100
		maxEvict := 100
		lru := NewGDSFCache(size, maxEvict)

		lru.Put(key1, val1)

		if val, ok := lru.Get(key1); !ok || !bytes.Equal(val, val1) {
			t.Errorf("can't find key1 in cache")
		}

		if !lru.Has(key1) {
			t.Errorf("can't find key1 in cache")
		}

		if _, ok := lru.Get(key2); ok {
			t.Errorf("key2 should be in cache")
		}

		if lru.Has(key2) {
			t.Errorf("key2 should be in cache")
		}

		lru.Put(key2, val2)
		if val, ok := lru.Get(key2); !ok || !bytes.Equal(val, val2) {
			t.Errorf("couldn't find key2 in cache or value is invalid")
		}

		lru.Put(key3, val3)

		if lru.Has(key1) {
			t.Errorf("key1 is still there")
		}

		expected_size := 70
		actual_size := lru.Size()
		if expected_size != actual_size {
			t.Errorf("expected cache size %d, but got %d", expected_size, actual_size)
		}

	})

	t.Run("priority queue check, same count", func(t *testing.T) {
		size := 100
		maxEvict := 100
		lru := NewGDSFCache(size, maxEvict)

		lru.Put(key1, val1)
		lru.Get(key1)
		lru.Get(key1)
		lru.Put(key2, val2)
		lru.Get(key2)
		lru.Get(key2)

		lru.Put(key3, val3)

		// since both key1 and key2 have a count of 3 and key2 was
		// most recently accessed, key1 should be evicted
		if lru.Has(key1) {
			t.Errorf("key1 wasn't evicted")
		}
	})

	t.Run("priority queue check, lower count is evicted", func(t *testing.T) {
		size := 100
		maxEvict := 100
		lru := NewGDSFCache(size, maxEvict)

		lru.Put(key1, val1)
		lru.Get(key1)
		lru.Get(key1)
		lru.Get(key1)
		lru.Put(key2, val2)
		lru.Get(key2)

		lru.Put(key3, val3)

		// key2 is evicted even though most recently accessed.
		// this is because key2 has higher count
		if lru.Has(key2) {
			t.Errorf("key2 wasn't evicted")
		}
	})

	t.Run("object is too big", func(t *testing.T) {
		// size := 100
		// maxEvict := 100
		// lru := NewGDSFCache(size, maxEvict)

		// lru.Put(key1, val1)
		// lru.Get(key1)
		// lru.Get(key1)
		// lru.Get(key1)
		// lru.Put(key2, val2)
		// lru.Get(key2)

		// lru.Put(key3, val3)

		// // key2 is evicted even though most recently accessed.
		// // this is because key2 has higher count
		// if lru.Has(key2) {
		// 	t.Errorf("key2 wasn't evicted")
		// }
	})

	t.Run("evict more than one key", func(t *testing.T) {
		// size := 100
		// maxEvict := 100
		// lru := NewGDSFCache(size, maxEvict)

		// lru.Put(key1, val1)
		// lru.Get(key1)
		// lru.Get(key1)
		// lru.Get(key1)
		// lru.Put(key2, val2)
		// lru.Get(key2)

		// lru.Put(key3, val3)

		// // key2 is evicted even though most recently accessed.
		// // this is because key2 has higher count
		// if lru.Has(key2) {
		// 	t.Errorf("key2 wasn't evicted")
		// }
	})

	t.Run("hit max evictions", func(t *testing.T) {
		// size := 100
		// maxEvict := 100
		// lru := NewGDSFCache(size, maxEvict)

		// lru.Put(key1, val1)
		// lru.Get(key1)
		// lru.Get(key1)
		// lru.Get(key1)
		// lru.Put(key2, val2)
		// lru.Get(key2)

		// lru.Put(key3, val3)

		// // key2 is evicted even though most recently accessed.
		// // this is because key2 has higher count
		// if lru.Has(key2) {
		// 	t.Errorf("key2 wasn't evicted")
		// }
	})
}
