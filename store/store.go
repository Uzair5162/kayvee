package store

import (
	"fmt"
	"sync"
)

type Store struct {
	mu		sync.RWMutex
	data	map[string]string
}

func New() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Set(k string, v string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[k] = v
}

func (s *Store) Get(k string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[k]
	return v, ok
}

func (s *Store) Del(k string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[k]; ok {
		delete(s.data, k)
		return true
	}
	return false
}

func (s *Store) Display() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for key, value := range s.data {
    	fmt.Println("Key:", key, "Value:", value)
	}
}
