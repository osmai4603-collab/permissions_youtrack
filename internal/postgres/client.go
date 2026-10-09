package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool initializes, validates, and returns a new pgxpool.Pool configured according to cfg.
func NewPool(ctx context.Context, cfg Config, logger *slog.Logger) (*pgxpool.Pool, error) {
	const op = "postgres.NewPool"

	if logger == nil {
		logger = slog.Default()
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, TranslateError(op, fmt.Errorf("failed to parse postgres DSN: %w", err))
	}

	// Apply lifecycle and connection pool options
	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	if cfg.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.HealthPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.HealthPeriod
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, TranslateError(op, fmt.Errorf("failed to create pgxpool instance: %w", err))
	}

	// Verify database connectivity
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, TranslateError(op, fmt.Errorf("failed to connect to database (ping): %w", err))
	}

	logger.InfoContext(ctx, "successfully connected to postgresql database",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", cfg.Database),
		slog.Int("max_conns", int(cfg.MaxConns)),
	)

	return pool, nil
}
