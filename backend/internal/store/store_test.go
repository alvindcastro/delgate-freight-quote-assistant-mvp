package store

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/quote"
)

func TestMemoryStoreEvictsOldestQuoteWhenLimitExceeded(t *testing.T) {
	store := NewMemoryStoreWithLimit(2)
	base := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)

	store.Save(quote.Response{QuoteID: "Q-1", CreatedAt: base})
	store.Save(quote.Response{QuoteID: "Q-2", CreatedAt: base.Add(time.Minute)})
	store.Save(quote.Response{QuoteID: "Q-3", CreatedAt: base.Add(2 * time.Minute)})

	if _, ok := store.Get("Q-1"); ok {
		t.Fatalf("expected oldest quote to be evicted")
	}
	if _, ok := store.Get("Q-2"); !ok {
		t.Fatalf("expected second quote to remain")
	}
	if _, ok := store.Get("Q-3"); !ok {
		t.Fatalf("expected newest quote to remain")
	}

	quotes := store.List()
	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}
	if quotes[0].QuoteID != "Q-3" || quotes[1].QuoteID != "Q-2" {
		t.Fatalf("expected newest-first order, got %q then %q", quotes[0].QuoteID, quotes[1].QuoteID)
	}
}

func TestMemoryStoreConcurrentSaveStaysBounded(t *testing.T) {
	store := NewMemoryStoreWithLimit(25)
	base := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			store.Save(quote.Response{
				QuoteID:   fmt.Sprintf("Q-%d", i),
				CreatedAt: base.Add(time.Duration(i) * time.Second),
			})
		}(i)
	}
	wg.Wait()

	if got := len(store.List()); got != 25 {
		t.Fatalf("expected bounded history of 25 quotes, got %d", got)
	}
}
