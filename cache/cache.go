package cache

type Cache interface {
	Get(key string) byte
	Put(key string, value []byte) byte
	Size() int
	Has(key string) bool
}
