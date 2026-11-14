package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/faber-numeris/sem/internal/infrastructure/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type stubPool struct{}

func (s *stubPool) Close() {}
func (s *stubPool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (s *stubPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (s *stubPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row { return nil }
func (s *stubPool) Ping(ctx context.Context) error                                { return nil }

func TestConnect_SuccessAndDefaultNewPoolFn(t *testing.T) {
	// Exercise default newPoolFn (pgxpool.NewWithConfig is lazy and does not dial immediately)
	cfg := &config.AppConfig{
		DatabaseConfig: config.DatabaseConfig{
			DBHost_:     "localhost",
			DBPort_:     5432,
			DBUser_:     "postgres",
			DBPassword_: "postgres",
			DBName_:     "postgres",
			DBSSLMode_:  "disable",
		},
	}
	pool, err := Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool == nil {
		t.Fatal("expected pool, got nil")
	}
	if GetPool() != pool {
		t.Fatal("expected GetPool to match returned pool")
	}
	pool.Close()
}

func TestConnect_ParseErrorAndPoolError(t *testing.T) {
	// Invalid sslmode triggers pgxpool.ParseConfig error
	badCfg := &config.AppConfig{
		DatabaseConfig: config.DatabaseConfig{
			DBHost_:     "localhost",
			DBPort_:     5432,
			DBUser_:     "postgres",
			DBPassword_: "postgres",
			DBName_:     "postgres",
			DBSSLMode_:  "invalid-ssl-mode",
		},
	}
	if _, err := Connect(context.Background(), badCfg); err == nil {
		t.Fatal("expected parse error for invalid sslmode")
	}

	orig := newPoolFn
	defer func() { newPoolFn = orig }()

	newPoolFn = func(ctx context.Context, poolCfg *pgxpool.Config) (AbstractPool, error) {
		return nil, errors.New("pool init failed")
	}
	goodCfg := &config.AppConfig{
		DatabaseConfig: config.DatabaseConfig{
			DBHost_:     "localhost",
			DBPort_:     5432,
			DBUser_:     "postgres",
			DBPassword_: "postgres",
			DBName_:     "postgres",
			DBSSLMode_:  "disable",
		},
	}
	if _, err := Connect(context.Background(), goodCfg); err == nil {
		t.Fatal("expected error from newPoolFn")
	}

	stub := &stubPool{}
	newPoolFn = func(ctx context.Context, poolCfg *pgxpool.Config) (AbstractPool, error) {
		return stub, nil
	}
	got, err := Connect(context.Background(), goodCfg)
	if err != nil || got != stub {
		t.Fatalf("expected stub pool, got %v, %v", got, err)
	}
}
