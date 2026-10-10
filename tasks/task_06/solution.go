package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	realCapacity := capacity
	if capacity <= 0 {
		realCapacity = 0
	}

	return &LRUCache[K, V]{
		items: make(map[K]*list.Element, realCapacity),
		capacity: realCapacity,
		ll: list.List{},
	}
}

func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	el, found := c.items[key]

	if !found {
        var zero V
        return zero, false
    }

    c.ll.MoveToFront(el)
    e := el.Value.(*entry[K, V])
    return e.value, true
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity > 0 {
		el, ok := c.items[key]

		if ok {
			e := el.Value.(*entry[K, V])
			e.value = value
			return
		} else if c.ll.Len() >= c.capacity {
			last := c.ll.Back()
			if last != nil {
    			e := last.Value.(*entry[K, V])
    			delete(c.items, e.key)
    			c.ll.Remove(last)
			}

			e := &entry[K, V]{key: key, value: value}
			el := c.ll.PushFront(e)
			c.items[key] = el
		} else {
			e := &entry[K, V]{key: key, value: value}
			el := c.ll.PushFront(e)
			c.items[key] = el
		}
	}
}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
