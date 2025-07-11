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
		cfg.EvictionInterval = time.Duration(1) * time.Second
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
	defer s.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = s.now().Add(time.Duration(ttl) * time.Second)
	}

	s.data[k] = Item{
		value: v,
		exp:   exp,
	}
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
	defer s.mu.Unlock()

	if _, ok := s.data[k]; ok {
		delete(s.data, k)
		s.persist()
		return true
	}
	return false
}

func (s *Store) persist() {
	if s.persister == nil {
		return
	}

	records := make(map[string]persistence.Record)
	for k, i := range s.data {
		records[k] = persistence.Record{
			Value: i.value,
			Exp:   i.exp,
		}
	}

	if err := s.persister.Save(context.Background(), records); err != nil {
		fmt.Println("persistence failed:", err)
	}
}

func (s *Store) Display() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k, i := range s.data {
		fmt.Println("Key:", k, "Value:", i.value)
	}
}
