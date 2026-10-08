package main

type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	realCapacity := capacity
	if capacity <= 0 {
		realCapacity = 0
	}

	return &Cache[K, V]{
		items:    make(map[K]V, realCapacity),
		capacity: realCapacity,
	}
}
func (c *Cache[K, V]) Get(k K) (v V, ok bool) {
	v, ok = c.items[k]
	return v, ok
}
func (c *Cache[K, V]) Set(k K, v V) bool {
	if c.capacity == 0 {
		return false
	} else {
		_, ok := c.items[k]
		if !ok && len(c.items) >= c.capacity {
			return false
		} else {
			c.items[k] = v
			return true
		}
	}
}
