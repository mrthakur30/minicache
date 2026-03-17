package cache

import (
	"sync"
	"sync/atomic"
	"time"
)





type Cache struct {
	
	hits      uint64
	misses    uint64
	evictions uint64

	mu        sync.RWMutex
	items     map[string]Item
	done      chan struct{}
	interval  time.Duration
	closeOnce sync.Once
}

type Stats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
}

func (s Stats) HitRatio() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0
	}
	return float64(s.Hits) / float64(total)
}

func (c *Cache) Stats() Stats {
	return Stats{
		Hits:      atomic.LoadUint64(&c.hits),
		Misses:    atomic.LoadUint64(&c.misses),
		Evictions: atomic.LoadUint64(&c.evictions),
	}
}

type Item struct{
	Value any
	ExpiresAt time.Time
}

func New(opts ...Option) *Cache {
	c :=  &Cache{
		items : make(map[string]Item),
		done:     make(chan struct{}),
		interval: time.Second,
	}

	for _,opt := range opts {
		opt(c)
	}

	go c.startEviction()
	return c 
}

func (c *Cache) startEviction(){
    ticker := time.NewTicker(c.interval)

	defer ticker.Stop()

	for {
		select {
		case <- ticker.C:
			c.deleteExpired()
		case <- c.done:
			return
		}
	}
}

func (c *Cache) deleteExpired() {
	now := time.Now()
	var evicted uint64

	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if !item.ExpiresAt.IsZero() && now.After(item.ExpiresAt) {
			delete(c.items, key)
			evicted++
		}
	}

	if evicted > 0 {
		atomic.AddUint64(&c.evictions, evicted)
	}
}

func (c *Cache) Close() error {
	c.closeOnce.Do(func() {
        close(c.done)
    })
	return nil
}

func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time

	if ttl > 0 {
         expiresAt = time.Now().Add(ttl)
	}

    c.items[key] = Item{
		Value : value, 
		ExpiresAt: expiresAt, 
	}

	return nil
}

func (c *Cache) Get(key string) (any, error){
	c.mu.RLock()
	item, ok := c.items[key] 
	c.mu.RUnlock()

	if !ok {
		atomic.AddUint64(&c.misses, 1)
		return nil, ErrNotFound
	}

	if !item.ExpiresAt.IsZero() && time.Now().After(item.ExpiresAt) {
		c.mu.Lock()
		defer c.mu.Unlock()

		if cur, ok := c.items[key]; ok && !cur.ExpiresAt.IsZero() && time.Now().After(cur.ExpiresAt) {
			delete(c.items, key)
			atomic.AddUint64(&c.evictions, 1)
		}
		atomic.AddUint64(&c.misses, 1)
		return nil, ErrExpired
	}

	atomic.AddUint64(&c.hits, 1)
	return item.Value, nil
}

func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
    
	delete(c.items,key)

	return nil
}
