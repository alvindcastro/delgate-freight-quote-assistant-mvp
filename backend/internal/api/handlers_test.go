package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/assistant"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/quote"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/store"
)

func TestQuoteIDsAreUniqueForRapidConcurrentRequests(t *testing.T) {
	quoteStore := store.NewMemoryStoreWithLimit(100)
	server := NewServerWithConfig(quoteStore, assistant.NewService(assistant.Config{}), Config{})
	handler := server.Routes()

	const requestCount = 50
	ids := make(chan string, requestCount)
	errs := make(chan error, requestCount)

	var wg sync.WaitGroup
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			body, err := json.Marshal(validQuoteRequest(i))
			if err != nil {
				errs <- err
				return
			}
			req := httptest.NewRequest(http.MethodPost, "/api/quote", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusCreated {
				errs <- fmt.Errorf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
				return
			}

			var resp quote.Response
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				errs <- err
				return
			}
			ids <- resp.QuoteID
		}(i)
	}
	wg.Wait()
	close(errs)
	close(ids)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	seen := make(map[string]bool)
	for id := range ids {
		if id == "" {
			t.Fatalf("expected quote ID")
		}
		if seen[id] {
			t.Fatalf("duplicate quote ID generated: %s", id)
		}
		seen[id] = true
	}
	if len(seen) != requestCount {
		t.Fatalf("expected %d quote IDs, got %d", requestCount, len(seen))
	}
	if got := len(quoteStore.List()); got != requestCount {
		t.Fatalf("expected %d stored quotes, got %d", requestCount, got)
	}
}

func TestRequestBodyLimitReturnsPayloadTooLarge(t *testing.T) {
	server := NewServerWithConfig(store.NewMemoryStore(), assistant.NewService(assistant.Config{}), Config{
		MaxRequestBodyBytes: 24,
	})
	body := `{"text":"` + strings.Repeat("x", 64) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/parse-request", strings.NewReader(body))
	rec := httptest.NewRecorder()

	server.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d: %s", http.StatusRequestEntityTooLarge, rec.Code, rec.Body.String())
	}
}

func TestQuoteRejectsUnknownAndTrailingJSON(t *testing.T) {
	server := NewServerWithConfig(store.NewMemoryStore(), assistant.NewService(assistant.Config{}), Config{})
	handler := server.Routes()

	body, err := json.Marshal(map[string]any{
		"origin":      map[string]string{"city": "Vancouver", "province": "BC", "postalCode": "V6B 1A1"},
		"destination": map[string]string{"city": "Calgary", "province": "AB", "postalCode": "T2P 1J9"},
		"weightLbs":   500,
		"unexpected":  "field",
	})
	if err != nil {
		t.Fatal(err)
	}
	unknownReq := httptest.NewRequest(http.MethodPost, "/api/quote", bytes.NewReader(body))
	unknownRec := httptest.NewRecorder()
	handler.ServeHTTP(unknownRec, unknownReq)
	if unknownRec.Code != http.StatusBadRequest {
		t.Fatalf("expected unknown field to return %d, got %d", http.StatusBadRequest, unknownRec.Code)
	}

	validBody, err := json.Marshal(validQuoteRequest(1))
	if err != nil {
		t.Fatal(err)
	}
	trailingReq := httptest.NewRequest(http.MethodPost, "/api/quote", strings.NewReader(string(validBody)+"{}"))
	trailingRec := httptest.NewRecorder()
	handler.ServeHTTP(trailingRec, trailingReq)
	if trailingRec.Code != http.StatusBadRequest {
		t.Fatalf("expected trailing JSON to return %d, got %d", http.StatusBadRequest, trailingRec.Code)
	}
}

func TestAPITokenProtectsQuoteEndpointsButNotHealth(t *testing.T) {
	server := NewServerWithConfig(store.NewMemoryStore(), assistant.NewService(assistant.Config{}), Config{
		APIToken: "secret-token",
	})
	handler := server.Routes()

	healthReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	healthRec := httptest.NewRecorder()
	handler.ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected health without token to return %d, got %d", http.StatusOK, healthRec.Code)
	}

	blockedReq := httptest.NewRequest(http.MethodGet, "/api/quotes", nil)
	blockedRec := httptest.NewRecorder()
	handler.ServeHTTP(blockedRec, blockedReq)
	if blockedRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing token to return %d, got %d", http.StatusUnauthorized, blockedRec.Code)
	}

	allowedReq := httptest.NewRequest(http.MethodGet, "/api/quotes", nil)
	allowedReq.Header.Set("Authorization", "Bearer secret-token")
	allowedRec := httptest.NewRecorder()
	handler.ServeHTTP(allowedRec, allowedReq)
	if allowedRec.Code != http.StatusOK {
		t.Fatalf("expected bearer token to return %d, got %d", http.StatusOK, allowedRec.Code)
	}

	headerReq := httptest.NewRequest(http.MethodGet, "/api/quotes", nil)
	headerReq.Header.Set("X-API-Token", "secret-token")
	headerRec := httptest.NewRecorder()
	handler.ServeHTTP(headerRec, headerReq)
	if headerRec.Code != http.StatusOK {
		t.Fatalf("expected X-API-Token to return %d, got %d", http.StatusOK, headerRec.Code)
	}
}

func validQuoteRequest(i int) quote.Request {
	return quote.Request{
		CustomerName:  fmt.Sprintf("Customer %d", i),
		CustomerEmail: fmt.Sprintf("customer%d@example.com", i),
		Origin: quote.Location{
			City:       "Vancouver",
			Province:   "BC",
			PostalCode: "V6B 1A1",
		},
		Destination: quote.Location{
			City:       "Calgary",
			Province:   "AB",
			PostalCode: "T2P 1J9",
		},
		ShipmentType: "ltl_pallet",
		Pieces:       1,
		Pallets:      1,
		WeightLbs:    500,
		LengthIn:     48,
		WidthIn:      40,
		HeightIn:     42,
		ServiceLevel: "standard",
	}
}
