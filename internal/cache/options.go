package cache

import "time"

type Option func(*Cache)

func WithEvictionInterval(d time.Duration) Option {
    return func(c *Cache) {
        if d > 0 {
            c.interval = d
        }
    }
}