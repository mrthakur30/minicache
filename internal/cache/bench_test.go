package cache

import (
	"strconv"
	"sync/atomic"
	"testing"
)

const (
	keyspace = 1024
)

func makeKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = "key_" + strconv.Itoa(i)
	}
	return keys
}

func prefill( c *Cache, keys []string){
	for i,k := range keys {
        _ = c.Set(k,i,0)
	}
}

func runSerialMix(b *testing.B, writeEvery int){
    c := New()
	keys := makeKeys(keyspace)
	prefill(c,keys)

	b.ReportAllocs()
	b.ResetTimer()

	for i:= 0 ; i< b.N ; i++ {
		k := keys[i%len(keys)]

		if i%writeEvery == 0 {
			_ = c.Set(k,i,0)
		} else {
			_ , _ = c.Get(k)
		}
	}
}

func runParallelMix(b *testing.B, writeEvery int) {
    c := New()
    keys := makeKeys(keyspace)
    prefill(c, keys)

    var ctr uint64
    b.ReportAllocs()
    b.ResetTimer()

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            i := int(atomic.AddUint64(&ctr, 1) - 1)
            k := keys[i%len(keys)]

            if i%writeEvery == 0 {
                _ = c.Set(k, i, 0)
            } else {
                _, _ = c.Get(k)
            }
        }
    })
}


func BenchmarkCache_ReadHeavy(b *testing.B) {
    // 90% reads, 10% writes
    runSerialMix(b, 10)
}

func BenchmarkCache_Balanced(b *testing.B) {
    // 50% reads, 50% writes
    runSerialMix(b, 2)
}

func BenchmarkCache_WriteHeavy(b *testing.B) {
    // 100% writes baseline for worst-case lock contention
    runSerialMix(b, 1)
}

// ---- Parallel benchmarks ----

func BenchmarkCacheParallel_ReadHeavy(b *testing.B) {
    runParallelMix(b, 10)
}

func BenchmarkCacheParallel_Balanced(b *testing.B) {
    runParallelMix(b, 2)
}

func BenchmarkCacheParallel_WriteHeavy(b *testing.B) {
    runParallelMix(b, 1)
}