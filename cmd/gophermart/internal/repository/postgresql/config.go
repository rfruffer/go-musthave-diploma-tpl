package postgresql

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func InitDB(dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("error parsing database config: %w", err)
	}
	config.MaxConns = 10
	config.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	ctx := context.Background()

	schema := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			login TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			balance NUMERIC(18, 2) DEFAULT 0,
			withdrawn NUMERIC(18, 2) DEFAULT 0
		);`,

		`CREATE TABLE IF NOT EXISTS orders (
			number TEXT PRIMARY KEY,
			user_id UUID REFERENCES users(id),
			status TEXT NOT NULL,
			accrual NUMERIC(18, 2),
			uploaded_at TIMESTAMP DEFAULT now()
		);`,

		`CREATE TABLE IF NOT EXISTS withdrawals (
			id SERIAL PRIMARY KEY,
			user_id UUID REFERENCES users(id),
			order_number TEXT NOT NULL,
			amount NUMERIC(18, 2) NOT NULL,
			processed_at TIMESTAMP DEFAULT now()
		);`,
	}

	for _, stmt := range schema {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to initialize schema: %w", err)
		}
	}

	log.Println("Database connection established and schema initialized")
	return pool, nil
}

func CloseDB(pool *pgxpool.Pool) {
	if pool == nil {
		log.Println("No database connection to close")
		return
	}
	pool.Close()
	log.Println("Database connection closed")
}
