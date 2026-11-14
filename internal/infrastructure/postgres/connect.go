package postgres

import (
	"context"
	"fmt"
	"sync"

	"github.com/faber-numeris/sem/internal/infrastructure/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	Pool      AbstractPool
	mu        sync.Mutex
	newPoolFn = func(ctx context.Context, poolCfg *pgxpool.Config) (AbstractPool, error) {
		return pgxpool.NewWithConfig(ctx, poolCfg)
	}
)

func Connect(ctx context.Context, cfg config.IDatabaseConfig) (AbstractPool, error) {
	mu.Lock()
	defer mu.Unlock()

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost(),
		cfg.DBPort(),
		cfg.DBUser(),
		cfg.DBPassword(),
		cfg.DBName(),
		cfg.DBSSLMode(),
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2

	pool, err := newPoolFn(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	Pool = pool
	return pool, nil
}

func GetPool() AbstractPool {
	mu.Lock()
	defer mu.Unlock()
	return Pool
}
