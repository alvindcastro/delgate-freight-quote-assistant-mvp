package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/api"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/assistant"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/store"
)

type serverConfig struct {
	Port                string
	BindAddr            string
	Addr                string
	APIToken            string
	OpenAIAPIKey        string
	OpenAIModel         string
	OpenAIBaseURL       string
	QuoteHistoryLimit   int
	MaxRequestBodyBytes int64
	ReadTimeout         time.Duration
	ReadHeaderTimeout   time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
}

func main() {
	if err := loadLocalEnv(); err != nil {
		log.Printf("Skipping local .env load: %v", err)
	}

	config, err := configFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	quoteStore := store.NewMemoryStoreWithLimit(config.QuoteHistoryLimit)
	assistantService := assistant.NewService(assistant.Config{
		APIKey:  config.OpenAIAPIKey,
		Model:   config.OpenAIModel,
		BaseURL: config.OpenAIBaseURL,
		Timeout: 10 * time.Second,
	})

	apiServer := api.NewServerWithConfig(quoteStore, assistantService, api.Config{
		APIToken:            config.APIToken,
		MaxRequestBodyBytes: config.MaxRequestBodyBytes,
	})
	server := &http.Server{
		Addr:              config.Addr,
		Handler:           apiServer.Routes(),
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
	}

	log.Printf("DelGate Freight Quote Assistant API listening on http://%s", config.Addr)
	log.Printf("OpenAI enabled: %t", config.OpenAIAPIKey != "")
	log.Printf("API token required: %t", config.APIToken != "")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func loadLocalEnv() error {
	loaded := false
	for _, path := range []string{".env", "backend/.env"} {
		err := loadEnvFile(path)
		if err == nil {
			loaded = true
			continue
		}
		if !os.IsNotExist(err) {
			return err
		}
	}
	if loaded {
		return nil
	}
	return nil
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s:%d: expected KEY=VALUE", path, lineNumber)
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			return fmt.Errorf("%s:%d: empty key", path, lineNumber)
		}
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func configFromEnv() (serverConfig, error) {
	bindAddr := getenv("BACKEND_BIND_ADDR", "127.0.0.1")
	token := strings.TrimSpace(os.Getenv("API_TOKEN"))
	if !isLocalBindAddr(bindAddr) && token == "" {
		return serverConfig{}, fmt.Errorf("API_TOKEN is required when BACKEND_BIND_ADDR is %q", bindAddr)
	}

	port := getenv("PORT", "8080")
	config := serverConfig{
		Port:                port,
		BindAddr:            bindAddr,
		Addr:                net.JoinHostPort(bindAddr, port),
		APIToken:            token,
		OpenAIAPIKey:        os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:         getenv("OPENAI_MODEL", "gpt-4o-mini"),
		OpenAIBaseURL:       getenv("OPENAI_BASE_URL", "https://api.openai.com/v1/chat/completions"),
		QuoteHistoryLimit:   getenvInt("QUOTE_HISTORY_LIMIT", store.DefaultMaxQuotes),
		MaxRequestBodyBytes: int64(getenvInt("MAX_REQUEST_BODY_BYTES", int(api.DefaultMaxRequestBodyBytes))),
		ReadTimeout:         getenvDuration("HTTP_READ_TIMEOUT", 10*time.Second),
		ReadHeaderTimeout:   getenvDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
		WriteTimeout:        getenvDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:         getenvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
	}
	return config, nil
}

func isLocalBindAddr(bindAddr string) bool {
	host := strings.TrimSpace(bindAddr)
	if host == "" || strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
