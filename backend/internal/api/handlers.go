package api

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/assistant"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/parser"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/quote"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/store"
)

type Server struct {
	store     *store.MemoryStore
	assistant *assistant.Service
	config    Config
}

type Config struct {
	APIToken            string
	MaxRequestBodyBytes int64
	StaticDir           string
}

const DefaultMaxRequestBodyBytes int64 = 1 << 20

var quoteIDFallbackCounter uint64

func NewServer(store *store.MemoryStore, assistant *assistant.Service) *Server {
	return NewServerWithConfig(store, assistant, Config{})
}

func NewServerWithConfig(store *store.MemoryStore, assistant *assistant.Service, config Config) *Server {
	if config.MaxRequestBodyBytes <= 0 {
		config.MaxRequestBodyBytes = DefaultMaxRequestBodyBytes
	}
	return &Server{store: store, assistant: assistant, config: config}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/quote", s.requireAPIToken(s.handleQuote))
	mux.HandleFunc("/api/quotes", s.requireAPIToken(s.handleQuotes))
	mux.HandleFunc("/api/parse-request", s.requireAPIToken(s.handleParseRequest))
	if strings.TrimSpace(s.config.StaticDir) != "" {
		mux.HandleFunc("/", s.handleStatic)
	}
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
	if !s.decodeJSON(w, r, &req) {
		return
	}

	now := time.Now().UTC()
	quoteID := newQuoteID(now)
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
	if !s.decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}
	writeJSON(w, http.StatusOK, parser.Parse(req.Text))
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	staticDir := strings.TrimSpace(s.config.StaticDir)
	cleanPath := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if cleanPath == "." {
		cleanPath = ""
	}

	filePath := filepath.Join(staticDir, filepath.FromSlash(cleanPath))
	if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
		if strings.HasPrefix(cleanPath, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		http.ServeFile(w, r, filePath)
		return
	}
	if strings.HasPrefix(cleanPath, "assets/") || path.Ext(cleanPath) != "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	indexPath := filepath.Join(staticDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		writeError(w, http.StatusNotFound, "frontend build not found")
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, indexPath)
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) decodeJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return false
		}
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return false
	}
	return true
}

func (s *Server) requireAPIToken(next http.HandlerFunc) http.HandlerFunc {
	token := strings.TrimSpace(s.config.APIToken)
	if token == "" {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if validAPIToken(r, token) {
			next(w, r)
			return
		}
		writeError(w, http.StatusUnauthorized, "missing or invalid API token")
	}
}

func validAPIToken(r *http.Request, expected string) bool {
	provided := strings.TrimSpace(r.Header.Get("X-API-Token"))
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		provided = strings.TrimSpace(auth[len("bearer "):])
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func newQuoteID(now time.Time) string {
	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err == nil {
		return fmt.Sprintf("Q-%s-%s", now.Format("20060102-150405"), strings.ToUpper(hex.EncodeToString(randomBytes[:])))
	}
	counter := atomic.AddUint64(&quoteIDFallbackCounter, 1)
	return fmt.Sprintf("Q-%s-%d-%d", now.Format("20060102-150405"), now.UnixNano(), counter)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
