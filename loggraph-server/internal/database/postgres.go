package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pool.Ping: %w", err)
	}
	log.Println("[db] connected")
	return pool, nil
}

// RunMigrations reads and executes SQL files from the migrations directory.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".sql" {
			continue
		}
		path := filepath.Join(migrationsDir, e.Name())
		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", e.Name(), err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("exec %s: %w", e.Name(), err)
		}
		log.Printf("[db] migration applied: %s", e.Name())
	}
	return nil
}
