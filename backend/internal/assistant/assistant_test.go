package assistant

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/quote"
)

func TestGenerateReturnsValidatedOpenAIOutput(t *testing.T) {
	req, resp := testQuote()
	valid := Local(req, resp)

	service := newOpenAITestService(t, mustJSON(t, valid))

	got := service.Generate(context.Background(), req, resp)

	if got.Provider != "openai" {
		t.Fatalf("expected openai provider, got %q", got.Provider)
	}
	if got.Summary != valid.Summary {
		t.Fatalf("expected validated summary, got %q", got.Summary)
	}
	if got.CustomerMessage != valid.CustomerMessage {
		t.Fatalf("expected validated customer message, got %q", got.CustomerMessage)
	}
	if !reflect.DeepEqual(got.FollowUpQuestions, valid.FollowUpQuestions) {
		t.Fatalf("expected follow-up questions %v, got %v", valid.FollowUpQuestions, got.FollowUpQuestions)
	}
}

func TestGenerateFallsBackWhenOpenAIOutputIsMalformed(t *testing.T) {
	req, resp := testQuote()
	service := newOpenAITestService(t, "{not-json")

	got := service.Generate(context.Background(), req, resp)

	assertLocalFallback(t, got, req, resp)
}

func TestGenerateFallsBackWhenOpenAIOutputMissesRequiredField(t *testing.T) {
	req, resp := testQuote()
	valid := Local(req, resp)
	body := map[string]any{
		"summary":         valid.Summary,
		"customerMessage": valid.CustomerMessage,
	}
	service := newOpenAITestService(t, mustJSON(t, body))

	got := service.Generate(context.Background(), req, resp)

	assertLocalFallback(t, got, req, resp)
}

func TestGenerateFallsBackWhenOpenAIOutputChangesPrice(t *testing.T) {
	req, resp := testQuote()
	changed := Local(req, resp)
	expectedRange := estimatedRange(resp)
	changedRange := "CAD 10-20"
	changed.Summary = strings.Replace(changed.Summary, expectedRange, changedRange, 1)
	changed.CustomerMessage = strings.Replace(changed.CustomerMessage, expectedRange, changedRange, 1)
	service := newOpenAITestService(t, mustJSON(t, changed))

	got := service.Generate(context.Background(), req, resp)

	assertLocalFallback(t, got, req, resp)
}

func TestGenerateFallsBackWhenOpenAIOutputOmitsNonBindingDisclaimer(t *testing.T) {
	req, resp := testQuote()
	changed := Local(req, resp)
	changed.CustomerMessage = strings.Replace(changed.CustomerMessage, "not a binding quote", "ready for booking", 1)
	service := newOpenAITestService(t, mustJSON(t, changed))

	got := service.Generate(context.Background(), req, resp)

	assertLocalFallback(t, got, req, resp)
}

func TestGenerateFallsBackWhenOpenAIOutputChangesFollowUpQuestions(t *testing.T) {
	req, resp := testQuoteWithMissingFields()
	changed := Local(req, resp)
	changed.FollowUpQuestions = []string{"Can we book this immediately?"}
	service := newOpenAITestService(t, mustJSON(t, changed))

	got := service.Generate(context.Background(), req, resp)

	assertLocalFallback(t, got, req, resp)
}

func newOpenAITestService(t *testing.T, content string) *Service {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("expected Authorization header, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{
					"message": map[string]string{
						"role":    "assistant",
						"content": content,
					},
				},
			},
		}); err != nil {
			t.Fatalf("failed to encode fake OpenAI response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	return NewService(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
}

func testQuote() (quote.Request, quote.Response) {
	req := quote.Request{
		CustomerName: "Maya",
		Origin: quote.Location{
			City:       "Vancouver",
			Province:   "BC",
			PostalCode: "V5K0A1",
		},
		Destination: quote.Location{
			City:       "Calgary",
			Province:   "AB",
			PostalCode: "T2P1J9",
		},
		ShipmentType: "ltl_pallet",
		Pallets:      2,
		WeightLbs:    1200,
		LengthIn:     48,
		WidthIn:      40,
		HeightIn:     48,
		ServiceLevel: "standard",
		Accessorials: []string{"liftgate"},
	}
	resp := quote.Calculate(req, quote.CalculateOptions{
		QuoteID:   "Q-TEST",
		CreatedAt: time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC),
	})
	return req, resp
}

func testQuoteWithMissingFields() (quote.Request, quote.Response) {
	req, _ := testQuote()
	req.Destination.PostalCode = ""
	req.WeightLbs = 0
	resp := quote.Calculate(req, quote.CalculateOptions{
		QuoteID:   "Q-MISSING",
		CreatedAt: time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC),
	})
	return req, resp
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return string(encoded)
}

func assertLocalFallback(t *testing.T, got Output, req quote.Request, resp quote.Response) {
	t.Helper()

	want := Local(req, resp)
	if got.Provider != "local-fallback" {
		t.Fatalf("expected local-fallback provider, got %q", got.Provider)
	}
	if got.Summary != want.Summary {
		t.Fatalf("expected local fallback summary %q, got %q", want.Summary, got.Summary)
	}
	if got.CustomerMessage != want.CustomerMessage {
		t.Fatalf("expected local fallback customer message %q, got %q", want.CustomerMessage, got.CustomerMessage)
	}
	if !reflect.DeepEqual(got.FollowUpQuestions, want.FollowUpQuestions) {
		t.Fatalf("expected local fallback questions %v, got %v", want.FollowUpQuestions, got.FollowUpQuestions)
	}
}
