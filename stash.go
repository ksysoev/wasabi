package wasabi

import "sync"

type Stasher interface {
	Set(key string, value any)
	Get(key string) any
	Delete(key string)
}

type StashStore struct {
	data map[string]any
	mu   sync.RWMutex
}

func NewStashStore() *StashStore {
	return &StashStore{data: make(map[string]any)}
}

func (s *StashStore) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *StashStore) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

func (s *StashStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}
