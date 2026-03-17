package main

import (
	"flag"
	"fmt"
	"strconv"
	"sync"
	"time"

	"minicache/internal/cache"
)

func main() {
	workers := flag.Int("workers", 8, "number of worker goroutines")
	ops := flag.Int("ops", 5000, "operations per worker")
	ttlMs := flag.Int("ttl", 50, "ttl in milliseconds for writes")
	evictionMs := flag.Int("eviction", 20, "eviction interval in milliseconds")
	flag.Parse()

	ttl := time.Duration(*ttlMs) * time.Millisecond
	evictionInterval := time.Duration(*evictionMs) * time.Millisecond

	c := cache.New(cache.WithEvictionInterval(evictionInterval))
	defer c.Close()

	var wg sync.WaitGroup
	start := time.Now()

	for worker := 0; worker < *workers; worker++ {
		worker := worker
		wg.Add(1)
		go func() {
			defer wg.Done()
			for op := 0; op < *ops; op++ {
				keyID := (worker*97 + (op / 5)) % 1024
				key := "k:" + strconv.Itoa(keyID)

				if op%5 == 0 {
					_ = c.Set(key, op, ttl)
				} else {
					_, _ = c.Get(key)
				}

				if op%20 == 0 {
					time.Sleep(ttl / 4)
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	s := c.Stats()
	fmt.Printf("workers=%d ops/worker=%d total_ops=%d ttl=%s eviction=%s\n", *workers, *ops, (*workers)*(*ops), ttl, evictionInterval)
	fmt.Printf("hits=%d misses=%d evictions=%d hit_ratio=%.3f elapsed=%s\n", s.Hits, s.Misses, s.Evictions, s.HitRatio(), elapsed)
}
