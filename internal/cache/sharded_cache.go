package cache

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"
)

type shard struct {
	mu    sync.RWMutex
	items map[string]Item
}

type ShardedCache struct {
	hits      uint64
	misses    uint64
	evictions uint64

	shards    []shard
	done      chan struct{}
	interval  time.Duration
	closeOnce sync.Once
}

func (c *ShardedCache) Stats() Stats {
	return Stats{
		Hits:      atomic.LoadUint64(&c.hits),
		Misses:    atomic.LoadUint64(&c.misses),
		Evictions: atomic.LoadUint64(&c.evictions),
	}
}

type ShardedOption func(*ShardedCache)

func WithShardedEvictionInterval(d time.Duration) ShardedOption {
	return func(c *ShardedCache) {
		if d > 0 {
			c.interval = d
		}
	}
}

func NewSharded(shardCount int, opts ...ShardedOption) *ShardedCache {
	if shardCount <= 0 {
		shardCount = 16
	}

	c := &ShardedCache{
		shards:   make([]shard, shardCount),
		done:     make(chan struct{}),
		interval: time.Second,
	}

	for i := range c.shards {
		c.shards[i].items = make(map[string]Item)
	}

	for _, opt := range opts {
		opt(c)
	}

	go c.startEviction()

	return c
}

func (c *ShardedCache) startEviction() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.done:
			return

		}
	}
}

func (c *ShardedCache) deleteExpired() {
	now := time.Now()
	var evicted uint64

	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()

		for key, item := range s.items {
			if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
				delete(s.items, key)
				evicted++
			}
		}
		s.mu.Unlock()
	}

	if evicted > 0 {
		atomic.AddUint64(&c.evictions, evicted)
	}
}

func (c *ShardedCache) Close() error {
	c.closeOnce.Do(func() {
		close(c.done)
	})
	return nil
}

func (c *ShardedCache) getShard(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	hash := h.Sum32()

	return &c.shards[hash%uint32(len(c.shards))]
}

func (c *ShardedCache) Set(key string, value any, ttl time.Duration) error {
    s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiresAt time.Time

	if ttl > 0 {
         expiresAt = time.Now().Add(ttl)
	}

    s.items[key] = Item{
		Value : value, 
		ExpiresAt: expiresAt, 
	}

	return nil
}

func (c *ShardedCache) Get(key string) (any, error) {
	s := c.getShard(key)
	s.mu.RLock()
	item, ok := s.items[key]
	s.mu.RUnlock()

	if !ok {
		atomic.AddUint64(&c.misses, 1)
		return nil, ErrNotFound
	}

	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		s.mu.Lock()
		defer s.mu.Unlock()

		if cur, ok := s.items[key]; ok && !cur.ExpiresAt.IsZero() && time.Now().After(cur.ExpiresAt) {
			delete(s.items, key)
			atomic.AddUint64(&c.evictions, 1)
		}
		atomic.AddUint64(&c.misses, 1)
		return nil, ErrExpired
	}

	atomic.AddUint64(&c.hits, 1)
	return item.Value, nil
}

func (c *ShardedCache) Delete(key string) error {
    s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()
    
	delete(s.items,key)

	return nil
}
