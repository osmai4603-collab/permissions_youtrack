package postgres

import (
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConfigDSN(t *testing.T) {
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		User:     "user@example.com",
		Password: "p@ss:/word?",
		Database: "youtrack",
		SSLMode:  "verify-full",
	}

	parsed, err := url.Parse(cfg.DSN())
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}
	password, hasPassword := parsed.User.Password()
	if !hasPassword || parsed.User.Username() != cfg.User || password != cfg.Password {
		t.Fatalf("DSN user credentials were not encoded correctly")
	}
	if parsed.Host != "localhost:5432" || parsed.Path != "/youtrack" {
		t.Fatalf("unexpected DSN host/path: %q %q", parsed.Host, parsed.Path)
	}
	if got := parsed.Query().Get("sslmode"); got != cfg.SSLMode {
		t.Fatalf("unexpected sslmode: got %q, want %q", got, cfg.SSLMode)
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		t.Fatalf("pgxpool rejected DSN: %v", err)
	}
	if poolConfig.ConnConfig.User != cfg.User || poolConfig.ConnConfig.Password != cfg.Password ||
		poolConfig.ConnConfig.Database != cfg.Database {
		t.Fatal("pgxpool parsed unexpected connection settings")
	}
}

func TestDefaultConfigReadsEnvironment(t *testing.T) {
	t.Setenv("DB_HOST", "db.example.test")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_USER", "app")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "permissions")
	t.Setenv("DB_SSLMODE", "require")

	cfg := DefaultConfig()
	if cfg.Host != "db.example.test" || cfg.Port != 6543 ||
		cfg.User != "app" || cfg.Password != "secret" ||
		cfg.Database != "permissions" || cfg.SSLMode != "require" {
		t.Fatalf("DefaultConfig did not load database environment variables: %+v", cfg)
	}
}
