# MiniCache

A concurrent in-memory cache built in Go to learn low-level backend engineering patterns:
- lock-based concurrency control
- TTL-based expiration
- background eviction lifecycle
- option-based constructors
- sharded cache for contention reduction
- benchmark-driven comparison
- atomic stats and CLI workload simulation

## What Is Implemented

### 1) `Cache` (single-map)
- `map[string]Item` protected by `sync.RWMutex`
- API:
  - `Set(key string, value any, ttl time.Duration) error`
  - `Get(key string) (any, error)`
  - `Delete(key string) error`
  - `Close() error`
  - `Stats() Stats`
- TTL semantics:
  - `ttl <= 0` => no expiration
  - `ttl > 0` => key expires at `time.Now().Add(ttl)`
- Lazy expiration in `Get` with re-check under write lock (avoids deleting fresh values written concurrently)
- Background eviction goroutine with ticker + `done` channel
- Safe shutdown via `sync.Once`

### 2) `ShardedCache` (multi-map)
- N shards (`[]shard`), each shard has its own `RWMutex + map`
- Key routing via FNV-1a hash (`hash % shardCount`)
- Same API behavior as `Cache`
- Independent shard locking reduces parallel contention

### 3) Options
- `WithEvictionInterval(d time.Duration)` for `Cache`
- `WithShardedEvictionInterval(d time.Duration)` for `ShardedCache`

### 4) Stats / Observability
- Atomic counters:
  - `Hits`
  - `Misses`
  - `Evictions`
- `HitRatio()` helper on `Stats`

### 5) CLI Demo
- `cmd/main.go` runs concurrent workers with configurable:
  - workers
  - ops per worker
  - TTL
  - eviction interval
- Prints final hit/miss/eviction counters and elapsed time

## Project Structure

```text
minicache/
  go.mod
  README.md
  cmd/
    main.go
  internal/
    cache/
      cache.go
      sharded_cache.go
      options.go
      errors.go
      cache_test.go
      bench_test.go
```

## API Contracts

### Errors
- `ErrNotFound`
- `ErrExpired`

### Core Types
- `Item { Value any; ExpiresAt time.Time }`
- `Stats { Hits, Misses, Evictions uint64 }`

## Concurrency / Correctness Notes

1. Go maps are not thread-safe; all map mutation is lock-protected.
2. Expired-key deletion in `Get` uses a write-lock re-check to avoid stale-read delete races.
3. Eviction worker is stoppable (`Close`) and idempotent (`sync.Once`).
4. Stats use atomics to avoid lock inflation on hot read paths.

## Run

From `minicache/`:

```bash
go run ./cmd
```

Custom run:

```bash
go run ./cmd -workers 8 -ops 5000 -ttl 50 -eviction 20
```

## Test

```bash
go test ./internal/cache/... -v
```

Run stats-only tests:

```bash
go test ./internal/cache/... -run "TestStats|TestShardedStats" -v
```

## Benchmark

```bash
go test ./internal/cache/... -run ^$ -bench . -benchmem
```

Shard-count sensitivity:

```bash
go test ./internal/cache/... -run ^$ -bench BenchmarkShards -benchmem
```

## Benchmark Summary (current)

Observed trend from the implemented benchmark suite:
- Single-map cache is competitive in serial paths.
- Sharded cache is significantly faster in parallel workloads due to reduced lock contention.
- Balanced parallel workloads show the biggest practical gains from sharding.

## Why This Project Matters

This project demonstrates the full learning arc from correctness to performance:
1. make concurrent access safe
2. add TTL semantics and lifecycle management
3. instrument behavior (stats)
4. measure with benchmarks
5. optimize with sharding based on measured contention

It is intentionally implementation-focused and testable, so every design claim can be verified with `go test` and `go test -bench`.
