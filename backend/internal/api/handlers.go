package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/assistant"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/parser"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/quote"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/store"
)

type Server struct {
	store     *store.MemoryStore
	assistant *assistant.Service
}

func NewServer(store *store.MemoryStore, assistant *assistant.Service) *Server {
	return &Server{store: store, assistant: assistant}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/quote", s.handleQuote)
	mux.HandleFunc("/api/quotes", s.handleQuotes)
	mux.HandleFunc("/api/parse-request", s.handleParseRequest)
	return cors(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}

func (s *Server) handleQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req quote.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	now := time.Now().UTC()
	quoteID := fmt.Sprintf("Q-%s", now.Format("20060102-150405"))
	resp := quote.Calculate(req, quote.CalculateOptions{QuoteID: quoteID, CreatedAt: now})

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
	defer cancel()
	aiOutput := s.assistant.Generate(ctx, req, resp)
	resp.AssistantSummary = aiOutput.Summary
	resp.CustomerMessage = aiOutput.CustomerMessage
	resp.FollowUpQuestions = aiOutput.FollowUpQuestions
	resp.AIProvider = aiOutput.Provider

	s.store.Save(resp)
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handleQuotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, s.store.List())
}

func (s *Server) handleParseRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req parser.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}
	writeJSON(w, http.StatusOK, parser.Parse(req.Text))
}

func cors(next http.Handler) http.Handler {
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == frontendOrigin || strings.HasPrefix(origin, "http://localhost:") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
		}
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
