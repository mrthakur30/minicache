package cache

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	c := New(WithEvictionInterval(20*time.Millisecond))

	if err := c.Set("name", "mukul", 0); err != nil {
		t.Fatalf("unexpected error from Set: %v", err)
	}

	val, err := c.Get("name")
	if err != nil {
		t.Fatalf("unexpected error from Get: %v", err)
	}

	if val != "mukul" {
		t.Fatalf("unexpected value: got=%v want=%v", val, "mukul")
	}
}

func TestGetMissing(t *testing.T) {
	c := New()

	_, err := c.Get("age")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestDelete(t *testing.T) {
	c := New()

	if err := c.Set("age", 56, 0); err != nil {
		t.Fatalf("unexpected error from Set: %v", err)
	}

	if err := c.Delete("age"); err != nil {
		t.Fatalf("unexpected error from Delete: %v", err)
	}

	_, err := c.Get("age")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got: %v", err)
	}
}

func TestOverwrite(t *testing.T) {
	c := New()

	if err := c.Set("age", 56, 0); err != nil {
		t.Fatalf("unexpected error from first Set: %v", err)
	}

	if err := c.Set("age", 45, 0); err != nil {
		t.Fatalf("unexpected error from second Set: %v", err)
	}

	val, err := c.Get("age")
	if err != nil {
		t.Fatalf("unexpected error from Get: %v", err)
	}

	if val != 45 {
		t.Fatalf("unexpected value: got=%v want=%v", val, 45)
	}
}

func TestConcurrent(t *testing.T) {
	c := New()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			_ = c.Set(key, "mukul", 0)
			_, _ = c.Get(key)
		}()
	}

	wg.Wait()
}

func TestGetExpired(t *testing.T) {
	c := New()

	if err := c.Set("token", "abc", 50*time.Millisecond); err != nil {
		t.Fatalf("unexpected error from Set: %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	_, err := c.Get("token")
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("expected ErrExpired, got: %v", err)
	}
}

func TestGetBeforeExpiry(t *testing.T) {
	c := New()

	if err := c.Set("token", "abc", 200*time.Millisecond); err != nil {
		t.Fatalf("unexpected error from Set: %v", err)
	}

	val, err := c.Get("token")
	if err != nil {
		t.Fatalf("unexpected error from Get: %v", err)
	}

	if val != "abc" {
		t.Fatalf("unexpected value: got=%v want=%v", val, "abc")
	}
}


func TestBackgroundEvictionRemovesExpiredItem(t *testing.T) {
    c := New()
    defer c.Close()

    if err := c.Set("token", "abc", 50*time.Millisecond); err != nil {
        t.Fatalf("set failed: %v", err)
    }

    time.Sleep(1200 * time.Millisecond) // enough for 1s ticker + expiry

    c.mu.RLock()
    _, ok := c.items["token"]
    c.mu.RUnlock()

    if ok {
        t.Fatalf("expected expired item to be evicted")
    }
}

func TestCloseTwiceIsSafe(t *testing.T) {
    c := New()
    if err := c.Close(); err != nil {
        t.Fatalf("first Close failed: %v", err)
    }
    if err := c.Close(); err != nil {
        t.Fatalf("second Close failed: %v", err)
    }
}

func TestBackgroundEvictionWithCustomInterval(t *testing.T) {
    c := New(WithEvictionInterval(20 * time.Millisecond))
    defer c.Close()

    if err := c.Set("k", "v", 30*time.Millisecond); err != nil {
        t.Fatalf("set failed: %v", err)
    }

    time.Sleep(120 * time.Millisecond)

    c.mu.RLock()
    _, ok := c.items["k"]
    c.mu.RUnlock()

    if ok {
        t.Fatalf("expected key to be evicted")
    }
}