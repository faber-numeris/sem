package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/faber-numeris/sem/internal/infrastructure/config"
	infrapostgres "github.com/faber-numeris/sem/internal/infrastructure/postgres"
	"github.com/faber-numeris/sem/internal/mocks"
	"github.com/stretchr/testify/mock"
)

func TestBootstrap_NewAppAndRun(t *testing.T) {
	ctx := context.Background()

	// 1. ProvideRepository / NewAppWithDeps error
	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	_, err = NewAppWithDeps(ctx, cfg, func(context.Context, config.IDatabaseConfig) (infrapostgres.AbstractPool, error) {
		return nil, errors.New("connect failed")
	})
	if err == nil {
		t.Fatal("expected error")
	}

	// 2. NewAppWithDeps success + repo.Ping check via /ready + Run graceful shutdown
	pool := mocks.NewMockAbstractPool(t)
	pool.EXPECT().Ping(mock.Anything).Return(nil).Once()
	pool.EXPECT().Close().Return().Once()

	t.Setenv("PORT", "0")
	cfgZero, err := config.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	app, err := NewAppWithDeps(ctx, cfgZero, func(context.Context, config.IDatabaseConfig) (infrapostgres.AbstractPool, error) {
		return pool, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	app.server.Handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	runCtx, cancel := context.WithCancel(ctx)
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	if err := app.Run(runCtx); err != nil {
		t.Fatal(err)
	}

	// 3. app.Run with ListenAndServe error (invalid address)
	poolErr := mocks.NewMockAbstractPool(t)
	poolErr.EXPECT().Close().Return().Once()
	appBadAddr := &App{
		server: &http.Server{Addr: "127.0.0.1:-1"},
		pool:   poolErr,
	}
	if err := appBadAddr.Run(ctx); err == nil {
		t.Fatal("expected server error")
	}

	// 4. NewApp / Run with config error
	t.Setenv("PORT", "not-a-port")
	if _, err := NewApp(ctx); err == nil {
		t.Fatal("expected config error")
	}
	if err := Run(ctx); err == nil {
		t.Fatal("expected config error from Run")
	}

	// 5. NewApp / Run with valid config and cancelled context
	t.Setenv("PORT", "0")
	cancelledCtx, cancelNow := context.WithCancel(ctx)
	cancelNow()
	if err := Run(cancelledCtx); err != nil {
		t.Fatal(err)
	}
}
