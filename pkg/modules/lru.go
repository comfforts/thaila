package modules

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"sync"
	"time"
)

const maxValueSize = 512 * 1024 * 1024 // 512MB

const (
	ERR_NOT_FOUND     = "error not found"
	ERR_LRU_CAP       = "error LRU capacity must be greater than zero"
	ERR_VAL_TOO_LARGE = "value size exceeds 512MB or set limit"
)

var (
	ErrNotFound      = errors.New(ERR_NOT_FOUND)
	ErrLRUCap        = errors.New(ERR_LRU_CAP)
	ErrValueTooLarge = errors.New(ERR_VAL_TOO_LARGE)
)

// LRUConfig holds configuration for LRUCache.
type LRUConfig struct {
	capacity int // Maximum number of items the cache can hold.
	limit    int // Maximum size of a single value in bytes (currently unused).
}

// NewLRUConfig returns a new LRUConfig with the given capacity.
func NewLRUConfig(capacity, limit int) LRUConfig {
	return LRUConfig{
		capacity: capacity,
		limit:    limit,
	}
}

// Node represents a doubly linked list node
type node[K comparable, V any] struct {
	key   K
	value V
	prev  *node[K, V]
	next  *node[K, V]
}

// LRUCache is a fixed-capacity Least Recently Used (LRU) cache.
// It is safe for concurrent use.
type LRUCache[K comparable, V any] struct {
	LRUConfig
	cache map[K]*node[K, V]
	head  *node[K, V]
	tail  *node[K, V]
	lock  sync.RWMutex
}

// NewLRUCache creates a new LRUCache with the specified configuration.
// Returns an error if the capacity is less than 1.
func NewLRUCache[K comparable, V any](cfg LRUConfig) (*LRUCache[K, V], error) {
	if cfg.capacity < 1 {
		return nil, ErrLRUCap
	}

	head := &node[K, V]{}
	tail := &node[K, V]{}
	head.next = tail
	tail.prev = head

	return &LRUCache[K, V]{
		LRUConfig: cfg,
		cache:     make(map[K]*node[K, V]),
		head:      head,
		tail:      tail,
	}, nil
}

// Set inserts or updates a value in the cache for the given key.
// If the cache exceeds its capacity, the least recently used item is evicted.
// Returns ErrValueTooLarge if the value is larger than limit or 512MB.
// The ttl parameter is currently unused.
func (c *LRUCache[K, V]) Set(ctx context.Context, key K, value V, ttl time.Duration) error {
	// Serialize value to check its size
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return err
	}
	if buf.Len() > maxValueSize {
		return ErrValueTooLarge
	}

	if c.limit > 0 && buf.Len() > c.limit {
		return ErrValueTooLarge
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if n, ok := c.cache[key]; ok {
		n.value = value
		c.moveToFront(n)
		return nil
	}

	if len(c.cache) >= c.capacity {
		lru := c.tail.prev
		c.remove(lru)
		delete(c.cache, lru.key)
	}

	n := &node[K, V]{key: key, value: value}
	c.cache[key] = n
	c.insertAtFront(n)
	return nil
}

// Get returns the value for the given key if present, or an error if not found.
// Accessing a key moves it to the most recently used position.
func (c *LRUCache[K, V]) Get(ctx context.Context, key K) (V, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if n, ok := c.cache[key]; ok {
		c.moveToFront(n)
		return n.value, nil
	}
	var zero V
	return zero, ErrNotFound
}

// Delete removes the value for the given key from the cache.
// Returns an error if the key is not present.
func (c *LRUCache[K, V]) Delete(ctx context.Context, key K) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if n, ok := c.cache[key]; ok {
		c.remove(n)
		delete(c.cache, key)
		return nil
	}
	return ErrNotFound
}

// Clear removes all entries from the cache and resets the linked list
func (c *LRUCache[K, V]) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Reset the map
	c.cache = make(map[K]*node[K, V])

	// Reset the linked list
	c.head.next = c.tail
	c.tail.prev = c.head
}

// Close doesn't do anything for LRU cache.
func (c *LRUCache[K, V]) Close() error {
	return nil
}

// moveToFront moves a node to front (MRU position)
func (c *LRUCache[K, V]) moveToFront(n *node[K, V]) {
	c.remove(n)
	c.insertAtFront(n)
}

func (c *LRUCache[K, V]) remove(n *node[K, V]) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

func (c *LRUCache[K, V]) insertAtFront(n *node[K, V]) {
	n.next = c.head.next
	n.prev = c.head
	c.head.next.prev = n
	c.head.next = n
}
