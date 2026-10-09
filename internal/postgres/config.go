package postgres

import (
	"net"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Config encapsulates database pool connection options and lifecycle parameters.
type Config struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"-"`
	Database        string        `json:"database"`
	SSLMode         string        `json:"ssl_mode"`
	MaxConns        int32         `json:"max_conns"`
	MinConns        int32         `json:"min_conns"`
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime"`
	HealthPeriod    time.Duration `json:"health_period"`
}

// DefaultConfig returns sensible defaults for connecting to PostgreSQL,
// reading from environment variables if present.
func DefaultConfig() Config {
	port := 5432
	if envPort := os.Getenv("DB_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	return Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            port,
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "postgres"),
		Database:        getEnv("DB_NAME", "youtrack"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxConns:        25,
		MinConns:        5,
		MaxConnIdleTime: 30 * time.Minute,
		MaxConnLifetime: 1 * time.Hour,
		HealthPeriod:    1 * time.Minute,
	}
}

// DSN formats the PostgreSQL connection URL string.
func (c Config) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   "/" + c.Database,
	}
	query := url.Values{}
	query.Set("sslmode", c.SSLMode)
	u.RawQuery = query.Encode()
	return u.String()
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
