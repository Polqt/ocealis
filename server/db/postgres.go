package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var Pool *pgxpool.Pool

func Connect() error {
	_ = godotenv.Load()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return fmt.Errorf("DATABASE_URL not set")
	}

	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("parse config error: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("pool creation error: %w", err)
	}

	Pool = pool
	fmt.Println("Postgres pool connected")
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
		fmt.Println("Postgres pool closed")
	}
}