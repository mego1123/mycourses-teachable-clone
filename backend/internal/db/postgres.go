package db

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"mycourses/internal/db/gen"
)

type DB struct {
	Pool    *pgxpool.Pool
	Queries *gen.Queries
}

type Config struct {
	URL         string
	MaxConns    int32
	MinConns    int32
	MaxConnIdle time.Duration
	MaxConnLife time.Duration
}

func New(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	if cfg.MaxConns == 0 { cfg.MaxConns = 25 }
	if cfg.MinConns == 0 { cfg.MinConns = 5 }
	if cfg.MaxConnIdle == 0 { cfg.MaxConnIdle = 5 * time.Minute }
	if cfg.MaxConnLife == 0 { cfg.MaxConnLife = 30 * time.Minute }

	pgxConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil { return nil, fmt.Errorf("failed to parse database URL: %w", err) }
	pgxConfig.MaxConns = cfg.MaxConns
	pgxConfig.MinConns = cfg.MinConns
	pgxConfig.MaxConnIdleTime = cfg.MaxConnIdle
	pgxConfig.MaxConnLifetime = cfg.MaxConnLife

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil { return nil, fmt.Errorf("failed to create connection pool: %w", err) }
	if err := pool.Ping(ctx); err != nil { pool.Close(); return nil, fmt.Errorf("failed to ping database: %w", err) }

	return &DB{Pool: pool, Queries: gen.New(pool)}, nil
}

func (db *DB) Close() { if db.Pool != nil { db.Pool.Close() } }

func (db *DB) Migrate(ctx context.Context, migrationsDir string) error {
	m, err := migrate.New(migrationsDir, db.Pool.Config().ConnString())
	if err != nil { return fmt.Errorf("failed to create migrate instance: %w", err) }
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange { return fmt.Errorf("failed to apply migrations: %w", err) }
	return nil
}

func (db *DB) WithTx(ctx context.Context, fn func(*gen.Queries) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil { return fmt.Errorf("failed to begin transaction: %w", err) }
	defer tx.Rollback(ctx)
	q := gen.New(tx)
	if err := fn(q); err != nil { return err }
	if err := tx.Commit(ctx); err != nil { return fmt.Errorf("failed to commit transaction: %w", err) }
	return nil
}

func (db *DB) HealthCheck(ctx context.Context) error { return db.Pool.Ping(ctx) }
