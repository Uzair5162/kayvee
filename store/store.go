package store

import (
	"context"
	"kayvee/persistence"
	"slices"
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

	done     chan struct{}
	stopOnce sync.Once

	now       func() time.Time
	persister persistence.Persister
}

type Config struct {
	EvictionInterval time.Duration
	Persister        persistence.Persister
}

func New(cfg Config) (*Store, error) {
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
		loaded, err := cfg.Persister.Load(context.Background())
		if err != nil {
			return nil, err
		}
		for k, r := range loaded {
			if r.Exp.IsZero() || r.Exp.After(s.now()) {
				s.data[k] = Item{value: r.Value, exp: r.Exp}
			}
		}
	}
	s.startEvictionLoop(cfg.EvictionInterval)
	return s, nil
}

func (s *Store) Set(k, v string, ttl int) error {
	s.mu.Lock()
	var exp time.Time
	if ttl > 0 {
		exp = s.now().Add(time.Second * time.Duration(ttl))
	}

	s.data[k] = Item{
		value: v,
		exp:   exp,
	}
	s.mu.Unlock()

	return s.persist()
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

func (s *Store) Del(k string) error {
	s.mu.Lock()
	_, ok := s.data[k]
	if ok {
		delete(s.data, k)
	}
	s.mu.Unlock()

	if ok {
		return s.persist()
	}
	return nil
}

func (s *Store) persist() error {
	if s.persister == nil {
		return nil
	}

	s.mu.RLock()
	records := make(map[string]persistence.Record, len(s.data))
	for k, i := range s.data {
		records[k] = persistence.Record{
			Value: i.value,
			Exp:   i.exp,
		}
	}
	s.mu.RUnlock()

	if err := s.persister.Save(context.Background(), records); err != nil {
		return err
	}
	return nil
}

func (s *Store) Shutdown() {
	s.stopOnce.Do(func() {
		close(s.done)
		_ = s.persist()
	})
}

func (s *Store) Snapshot() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	snapshot := make([]string, 0, len(keys))
	now := s.now()
	for _, k := range keys {
		i := s.data[k]
		line := k + ": " + i.value

		if !i.exp.IsZero() {
			ttl := i.exp.Sub(now)
			if ttl < 0 {
				ttl = 0
			}
			line += " exp in " + ttl.Truncate(time.Millisecond).String()
		}
		snapshot = append(snapshot, line)
	}
	return snapshot
}
