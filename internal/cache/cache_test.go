package cache

import (
    "errors"
    "fmt"
    "sync"
    "testing"
)

func TestSetAndGet(t *testing.T) {
    c := New()

    if err := c.Set("name", "mukul"); err != nil {
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

    if err := c.Set("age", 56); err != nil {
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

    if err := c.Set("age", 56); err != nil {
        t.Fatalf("unexpected error from first Set: %v", err)
    }

    if err := c.Set("age", 45); err != nil {
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
            _ = c.Set(key, "mukul")
            _, _ = c.Get(key)
        }()
    }

    wg.Wait()
}