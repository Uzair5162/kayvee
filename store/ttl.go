package store

import (
	"time"
)

func (s *Store) startEvictionLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				s.evictExpiredKeys()
			case <-s.done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Store) evictExpiredKeys() {
	s.mu.Lock()
	persist := false
	for k, i := range s.data {
		if !i.exp.IsZero() && i.exp.Before(s.now()) {
			delete(s.data, k)
			persist = true
		}
	}
	s.mu.Unlock()
	if persist {
		_ = s.persist()
	}
}
