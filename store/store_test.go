package store

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSetGet(t *testing.T) {
	s := newTestStore(t)

	if err := s.Set("k", "v", 0); err != nil {
		t.Fatal(err)
	}

	if got, ok := s.Get("k"); !ok || got != "v" {
		t.Fatalf("get failed, got: %q, ok=%v", got, ok)
	}

	if _, ok := s.Get("a"); ok {
		t.Fatal("expected key a not to exist")
	}
}

func TestSetGetTTL(t *testing.T) {
	s := newTestStore(t)

	fakeNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return fakeNow }

	if err := s.Set("k", "v", 1); err != nil {
		t.Fatal(err)
	}

	if got, ok := s.Get("k"); !ok || got != "v" {
		t.Fatalf("get failed, got: %q, ok=%v", got, ok)
	}

	s.now = func() time.Time { return fakeNow.Add(2 * time.Second) }

	if _, ok := s.Get("k"); ok {
		t.Fatal("expected key k to be expired and deleted")
	}
}

func TestEvictionTTL(t *testing.T) {
	s := newTestStore(t)

	fakeNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return fakeNow }

	if err := s.Set("k", "v", 1); err != nil {
		t.Fatal(err)
	}

	if _, ok := s.Get("k"); !ok {
		t.Fatal("expected key to exist before eviction")
	}

	s.now = func() time.Time { return fakeNow.Add(2 * time.Second) }
	s.evictExpiredKeys()

	s.mu.RLock()
	_, ok := s.data["k"]
	s.mu.RUnlock()
	if ok {
		t.Fatal("expected key to be evicted")
	}
}

func TestSetDel(t *testing.T) {
	s := newTestStore(t)

	if err := s.Set("k", "v", 0); err != nil {
		t.Fatal(err)
	}

	if _, ok := s.Get("k"); !ok {
		t.Fatal("expected key k to exist before deletion")
	}

	if err := s.Del("k"); err != nil {
		t.Fatal(err)
	}

	if _, ok := s.Get("k"); ok {
		t.Fatal("expected key k to not exist after deletion")
	}
}

func TestConcurrency(t *testing.T) {
	s := newTestStore(t)

	const goroutines = 50
	const operations = 1000

	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("key-%d", j%10)
				switch j % 3 {
				case 0:
					if err := s.Set(key, fmt.Sprintf("val-%d-%d", id, j), 0); err != nil {
						t.Errorf("set failed: %v", err)
					}
				case 1:
					_, _ = s.Get(key)
				default:
					_ = s.Del(key)
				}
			}
		}(i)
	}
	wg.Wait()

	if err := s.Set("k", "v", 0); err != nil {
		t.Fatalf("set failed after concurrent operations: %v", err)
	}
	if got, ok := s.Get("k"); !ok || got != "v" {
		t.Fatalf("get failed after concurrent operations, got: %q, ok=%v", got, ok)
	}
}
