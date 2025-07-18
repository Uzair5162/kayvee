package store

import (
	"context"
	"fmt"
	"kayvee/persistence"
	"sync"
	"time"
)

type Item struct {
	value string
	exp   time.Time
}

type Store struct {
	mu   sync.RWMutex
	data map[string]Item
	done chan struct{}

	now       func() time.Time
	persister persistence.Persister
}

type Config struct {
	EvictionInterval time.Duration
	Persister        persistence.Persister
}

func New(cfg Config) *Store {
	if cfg.EvictionInterval <= 0 {
		cfg.EvictionInterval = time.Second
	}
	s := &Store{
		data:      make(map[string]Item),
		done:      make(chan struct{}),
		now:       time.Now,
		persister: cfg.Persister,
	}

	if cfg.Persister != nil {
		if loaded, err := cfg.Persister.Load(context.Background()); err == nil {
			for k, r := range loaded {
				s.data[k] = Item{value: r.Value, exp: r.Exp}
			}
		} else {
			fmt.Println("failed to load data:", err)
		}
	}
	s.startEvictionLoop(cfg.EvictionInterval)
	return s
}

func (s *Store) Set(k string, v string, ttl int) {
	s.mu.Lock()
	var exp time.Time
	if ttl > 0 {
		exp = s.now().Add(time.Duration(ttl) * time.Second)
	}

	s.data[k] = Item{
		value: v,
		exp:   exp,
	}
	s.mu.Unlock()

	s.persist()
}

func (s *Store) Get(k string) (string, bool) {
	s.mu.RLock()
	i, ok := s.data[k]
	s.mu.RUnlock()

	if !ok {
		return "", false
	}

	if !i.exp.IsZero() && i.exp.Before(s.now()) {
		s.mu.Lock()
		delete(s.data, k)
		s.mu.Unlock()
		return "", false
	}

	return i.value, true
}

func (s *Store) Del(k string) bool {
	s.mu.Lock()
	_, ok := s.data[k]
	if ok {
		delete(s.data, k)
	}
	s.mu.Unlock()

	if ok {
		s.persist()
	}
	return ok
}

func (s *Store) persist() {
	if s.persister == nil {
		return
	}

	s.mu.RLock()
	records := make(map[string]persistence.Record)
	for k, i := range s.data {
		records[k] = persistence.Record{
			Value: i.value,
			Exp:   i.exp,
		}
	}
	s.mu.RUnlock()

	if err := s.persister.Save(context.Background(), records); err != nil {
		fmt.Println("persistence failed:", err)
	}
}

func (s *Store) Snapshot() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := make([]string, 0, len(s.data))
	for k, i := range s.data {
		line := k + ": " + i.value

		if !i.exp.IsZero() {
			ttl := i.exp.Sub(s.now())
			if ttl < 0 {
				ttl = 0
			}
			line += " exp in " + ttl.Truncate(time.Millisecond).String()
		}
		snapshot = append(snapshot, line)
	}
	return snapshot
}
