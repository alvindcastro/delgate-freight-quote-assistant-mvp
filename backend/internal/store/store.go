package store

import (
	"sort"
	"sync"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/quote"
)

type MemoryStore struct {
	mu     sync.RWMutex
	quotes map[string]quote.Response
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{quotes: make(map[string]quote.Response)}
}

func (s *MemoryStore) Save(resp quote.Response) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quotes[resp.QuoteID] = resp
}

func (s *MemoryStore) List() []quote.Response {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]quote.Response, 0, len(s.quotes))
	for _, value := range s.quotes {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].CreatedAt.After(values[j].CreatedAt)
	})
	return values
}

func (s *MemoryStore) Get(id string) (quote.Response, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.quotes[id]
	return value, ok
}
