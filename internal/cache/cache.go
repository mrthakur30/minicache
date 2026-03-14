package cache

import (
	"sync"
)

type Cache struct {
	mu sync.RWMutex
	items map[string]any
}


func New() *Cache {
	return &Cache{
		items : make(map[string]any),
	}
}

func (c *Cache) Set(key string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

    c.items[key] = value

	return nil
}

func (c *Cache) Get(key string) (any, error){
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.items[key] 

	if !ok {
		  return nil, ErrNotFound
	}

	return val, nil
}

func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
    
	delete(c.items,key)

	return nil
}
