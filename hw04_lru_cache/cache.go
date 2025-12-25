package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}
type CacheEntry struct {
	key   Key         // Ключ - нужен для удаления из словаря
	value interface{} // Значение
}

func (lruCache *lruCache) Set(key Key, value interface{}) bool {
	lruCache.mu.Lock()
	defer lruCache.mu.Unlock()

	var exists bool
	var v *ListItem
	cacheEntry := &CacheEntry{key: key, value: value}
	if v, exists = lruCache.items[key]; exists {
		v.Value = cacheEntry
		lruCache.queue.MoveToFront(v)
	} else {
		val := lruCache.queue.PushFront(cacheEntry)
		lruCache.items[key] = val
		if lruCache.queue.Len() > lruCache.capacity {
			back := lruCache.queue.Back()
			lruCache.queue.Remove(back)
			oldestEntry := back.Value.(*CacheEntry)
			delete(lruCache.items, oldestEntry.key)
		}
	}
	return exists
}

func (lruCache *lruCache) Get(key Key) (interface{}, bool) {
	lruCache.mu.Lock()
	defer lruCache.mu.Unlock()

	var exists bool
	var v *ListItem
	if v, exists = lruCache.items[key]; exists {
		lruCache.queue.MoveToFront(v)
		cacheEntry := v.Value.(*CacheEntry)
		return cacheEntry.value, exists
	}
	return nil, false
}

func (lruCache *lruCache) Clear() {
	lruCache.mu.Lock()
	defer lruCache.mu.Unlock()

	lruCache.items = make(map[Key]*ListItem, lruCache.capacity)
	lruCache.queue = NewList()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mu       sync.Mutex
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
