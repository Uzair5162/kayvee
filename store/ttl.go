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
	defer s.mu.Unlock()

	for k, i := range s.data {
		if !i.exp.IsZero() && i.exp.Before(time.Now()) {
			delete(s.data, k)
		}
	}
}

func (s *Store) StopEvictionLoop() {
	close(s.done)
}
