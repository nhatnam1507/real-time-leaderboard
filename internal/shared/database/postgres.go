// Package database provides database connection management.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"real-time-leaderboard/internal/config"
	"real-time-leaderboard/internal/shared/logger"
)

// Postgres represents a PostgreSQL connection pool
type Postgres struct {
	Pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewPostgres creates a new PostgreSQL connection pool
func NewPostgres(cfg config.DatabaseConfig, l *logger.Logger) (*Postgres, error) {
	dsn := cfg.GetDSN()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Validate and convert int to int32 with overflow protection
	if cfg.MaxConnections > 2147483647 || cfg.MaxConnections < 0 {
		return nil, fmt.Errorf("MaxConnections value %d is out of int32 range", cfg.MaxConnections)
	}
	poolConfig.MaxConns = int32(cfg.MaxConnections)

	if cfg.MaxIdleConns > 2147483647 || cfg.MaxIdleConns < 0 {
		return nil, fmt.Errorf("MaxIdleConns value %d is out of int32 range", cfg.MaxIdleConns)
	}
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	l.Info(context.Background(), "Database connection established")

	return &Postgres{
		Pool:   pool,
		logger: l,
	}, nil
}

// Close closes the database connection pool
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
		p.logger.Info(context.Background(), "Database connection closed")
	}
}

// Health checks the health of the database connection
func (p *Postgres) Health(ctx context.Context) error {
	return p.Pool.Ping(ctx)
}
