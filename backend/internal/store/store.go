package store

import (
	"sort"
	"sync"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/quote"
)

type MemoryStore struct {
	mu     sync.RWMutex
	quotes map[string]quote.Response
	order  []string
	limit  int
}

const DefaultMaxQuotes = 1000

func NewMemoryStore() *MemoryStore {
	return NewMemoryStoreWithLimit(DefaultMaxQuotes)
}

func NewMemoryStoreWithLimit(limit int) *MemoryStore {
	if limit <= 0 {
		limit = DefaultMaxQuotes
	}
	return &MemoryStore{
		quotes: make(map[string]quote.Response),
		order:  make([]string, 0, limit),
		limit:  limit,
	}
}

func (s *MemoryStore) Save(resp quote.Response) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.quotes[resp.QuoteID]; !exists {
		s.order = append(s.order, resp.QuoteID)
	}
	s.quotes[resp.QuoteID] = resp
	for len(s.order) > s.limit {
		oldest := s.order[0]
		s.order = s.order[1:]
		delete(s.quotes, oldest)
	}
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
