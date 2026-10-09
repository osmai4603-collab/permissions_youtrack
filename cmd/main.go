package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"youtrack/internal/postgres"
	"youtrack/internal/repositories"
	perms "youtrack/internal/services/permissions_services"
	"youtrack/migrations"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger.InfoContext(ctx, "starting YouTrack Permissions service")

	// 1. Load PostgreSQL configuration
	cfg := postgres.DefaultConfig()
	logger.InfoContext(ctx, "database configuration loaded",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", cfg.Database),
	)

	// 2. Initialize PostgreSQL connection pool
	pool, err := postgres.NewPool(ctx, cfg, logger)
	if err != nil {
		logger.ErrorContext(ctx, "failed to initialize postgres connection pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 3. Execute embedded database migrations
	logger.InfoContext(ctx, "running embedded database schema migrations")
	if err := postgres.RunMigrations(ctx, pool, migrations.FS, ".", logger); err != nil {
		logger.ErrorContext(ctx, "failed to execute database migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 4. Initialize health checker and ping pool
	healthChecker := postgres.NewHealthChecker(pool)
	if err := healthChecker.Check(ctx); err != nil {
		logger.ErrorContext(ctx, "health check ping failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.InfoContext(ctx, "database health check passed", slog.Any("pool_stats", healthChecker.Stats()))

	// 5. Initialize Services and Repositories
	permissionSvc := perms.NewService()
	rolesRepository := repositories.NewRolesService(permissionSvc, pool)

	// Quick smoke test on repository layer
	sysAdminRole, err := rolesRepository.GetRole(ctx, "SYSTEM_ADMIN")
	if err != nil {
		logger.WarnContext(ctx, "unable to fetch SYSTEM_ADMIN role", slog.String("error", err.Error()))
	} else {
		logger.InfoContext(ctx, "retrieved system admin role from postgres",
			slog.String("id", sysAdminRole.ID),
			slog.String("name", sysAdminRole.Name),
			slog.Int("permissions_count", len(sysAdminRole.Permissions)),
		)
	}

	logger.InfoContext(ctx, "YouTrack Permissions PostgreSQL backend ready")

	// Wait for shutdown signal or termination
	<-ctx.Done()
	logger.Info("shutting down YouTrack Permissions service gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = shutdownCtx
	logger.Info("service shutdown complete")
}
