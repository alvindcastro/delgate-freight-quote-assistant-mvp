package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/api"
	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/store"
)

func TestLoadEnvFileSetsUnsetValuesAndPreservesExistingValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := []byte(`
# Local backend config
DOTENV_TEST_FILE_VALUE=from-file
export DOTENV_TEST_EXPORTED="from-export"
DOTENV_TEST_EXISTING=from-file
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DOTENV_TEST_FILE_VALUE", "")
	t.Setenv("DOTENV_TEST_EXPORTED", "")
	t.Setenv("DOTENV_TEST_EXISTING", "from-env")

	if err := loadEnvFile(path); err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_TEST_FILE_VALUE"); got != "from-file" {
		t.Fatalf("expected dotenv value, got %q", got)
	}
	if got := os.Getenv("DOTENV_TEST_EXPORTED"); got != "from-export" {
		t.Fatalf("expected exported dotenv value without quotes, got %q", got)
	}
	if got := os.Getenv("DOTENV_TEST_EXISTING"); got != "from-env" {
		t.Fatalf("expected process env to win, got %q", got)
	}
}

func TestLoadLocalEnvLoadsBackendEnvFromRepoRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "backend"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "backend", ".env"), []byte("DOTENV_TEST_BACKEND_VALUE=from-backend\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DOTENV_TEST_BACKEND_VALUE", "")
	t.Chdir(root)

	if err := loadLocalEnv(); err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("DOTENV_TEST_BACKEND_VALUE"); got != "from-backend" {
		t.Fatalf("expected backend/.env value, got %q", got)
	}
}

func TestConfigFromEnvUsesSafeDefaults(t *testing.T) {
	clearConfigEnv(t)

	config, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	if config.Addr != "127.0.0.1:8080" {
		t.Fatalf("expected localhost default address, got %q", config.Addr)
	}
	if config.QuoteHistoryLimit != store.DefaultMaxQuotes {
		t.Fatalf("expected default quote history limit %d, got %d", store.DefaultMaxQuotes, config.QuoteHistoryLimit)
	}
	if config.MaxRequestBodyBytes != api.DefaultMaxRequestBodyBytes {
		t.Fatalf("expected default body limit %d, got %d", api.DefaultMaxRequestBodyBytes, config.MaxRequestBodyBytes)
	}
	if config.ReadTimeout != 10*time.Second {
		t.Fatalf("expected read timeout of 10s, got %s", config.ReadTimeout)
	}
	if config.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("expected read header timeout of 5s, got %s", config.ReadHeaderTimeout)
	}
	if config.WriteTimeout != 15*time.Second {
		t.Fatalf("expected write timeout of 15s, got %s", config.WriteTimeout)
	}
	if config.IdleTimeout != 60*time.Second {
		t.Fatalf("expected idle timeout of 60s, got %s", config.IdleTimeout)
	}
}

func TestConfigFromEnvRequiresTokenForPublicBind(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("BACKEND_BIND_ADDR", "0.0.0.0")

	if _, err := configFromEnv(); err == nil {
		t.Fatalf("expected public bind without API_TOKEN to fail")
	}
}

func TestConfigFromEnvAllowsExplicitPublicDemoWithoutToken(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("BACKEND_BIND_ADDR", "0.0.0.0")
	t.Setenv("ALLOW_PUBLIC_API", "true")
	t.Setenv("PORT", "18080")
	t.Setenv("STATIC_DIR", "/app/public")

	config, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	if config.Addr != "0.0.0.0:18080" {
		t.Fatalf("expected public bind address, got %q", config.Addr)
	}
	if !config.AllowPublicAPI {
		t.Fatalf("expected public API opt-in to be set")
	}
	if config.StaticDir != "/app/public" {
		t.Fatalf("expected static dir to be configured, got %q", config.StaticDir)
	}
	if config.APIToken != "" {
		t.Fatalf("expected no API token for public demo config")
	}
}

func TestConfigFromEnvAllowsPublicBindWithToken(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("BACKEND_BIND_ADDR", "0.0.0.0")
	t.Setenv("API_TOKEN", "secret-token")
	t.Setenv("PORT", "18080")

	config, err := configFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	if config.Addr != "0.0.0.0:18080" {
		t.Fatalf("expected public bind address, got %q", config.Addr)
	}
	if config.APIToken != "secret-token" {
		t.Fatalf("expected API token to be configured")
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"API_TOKEN",
		"ALLOW_PUBLIC_API",
		"BACKEND_BIND_ADDR",
		"HTTP_IDLE_TIMEOUT",
		"HTTP_READ_HEADER_TIMEOUT",
		"HTTP_READ_TIMEOUT",
		"HTTP_WRITE_TIMEOUT",
		"MAX_REQUEST_BODY_BYTES",
		"OPENAI_API_KEY",
		"OPENAI_BASE_URL",
		"OPENAI_MODEL",
		"PORT",
		"QUOTE_HISTORY_LIMIT",
	} {
		t.Setenv(key, "")
	}
}
