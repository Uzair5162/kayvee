package store

import (
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	s := New(Config{})

	s.Set("k", "v", 0)
	got, ok := s.Get("k")
	if !ok {
		t.Fatal("expected key k to exist")
	}
	if got != "v" {
		t.Errorf("expected value v, got %q", got)
	}

	_, ok = s.Get("a")
	if ok {
		t.Fatal("expected key a not to exist")
	}
}

func TestSetGetTTL(t *testing.T) {
	s := New(Config{})
	fakeNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return fakeNow }

	s.Set("k", "v", 1)
	got, ok := s.Get("k")
	if !ok {
		t.Fatal("expected key k to exist")
	}
	if got != "v" {
		t.Errorf("expected value v, got %q", got)
	}

	s.now = func() time.Time { return fakeNow.Add(2 * time.Second) }
	_, ok = s.Get("k")
	if ok {
		t.Fatal("expected key k to be expired and deleted")
	}
}

func TestEvictionTTL(t *testing.T) {
	s := New(Config{})
	fakeNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return fakeNow }

	s.Set("k", "v", 1)
	_, ok := s.Get("k")
	if !ok {
		t.Fatal("expected key to exist before eviction")
	}

	s.now = func() time.Time { return fakeNow.Add(2 * time.Second) }
	s.evictExpiredKeys()

	s.mu.RLock()
	_, ok = s.data["k"]
	s.mu.RUnlock()
	if ok {
		t.Fatal("expected key to be evicted")
	}
}

func TestSetDel(t *testing.T) {
	s := New(Config{})

	s.Set("k", "v", 0)
	_, ok := s.Get("k")
	if !ok {
		t.Fatal("expected key k to exist before deletion")
	}

	ok = s.Del("k")
	if !ok {
		t.Fatal("expected key k to be deleted")
	}

	_, ok = s.Get("k")
	if ok {
		t.Fatal("expected key k to not exist after deletion")
	}
}
