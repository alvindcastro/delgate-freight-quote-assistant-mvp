package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/api"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/assistant"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/store"
)

func main() {
	port := getenv("PORT", "8080")

	quoteStore := store.NewMemoryStore()
	assistantService := assistant.NewService(assistant.Config{
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   getenv("OPENAI_MODEL", "gpt-4o-mini"),
		BaseURL: getenv("OPENAI_BASE_URL", "https://api.openai.com/v1/chat/completions"),
		Timeout: 10 * time.Second,
	})

	server := api.NewServer(quoteStore, assistantService)

	addr := ":" + port
	log.Printf("DelGate Freight Quote Assistant API listening on http://localhost%s", addr)
	log.Printf("OpenAI enabled: %t", os.Getenv("OPENAI_API_KEY") != "")

	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
