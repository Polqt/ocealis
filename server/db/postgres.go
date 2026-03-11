package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

var Pool *pgxpool.Pool

func Connect(log *zap.Logger) error {
	_ = godotenv.Load()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}

	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 1 * time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("pool creation error: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("ping database:%w", err)
	}

	Pool = pool
	log.Info("database connected", zap.Int32("max_conns", cfg.MaxConns), zap.String("health_check_period", cfg.HealthCheckPeriod.String()))
	return nil
}

// WithTransaction executes the provided function within a database transaction.
// If the function returns an error, the transaction is rolled back;
// otherwise, it is committed.
func WithTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(q *ocealis.Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction:%w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(ocealis.New(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
