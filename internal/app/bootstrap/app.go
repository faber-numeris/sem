package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/faber-numeris/sem/internal/adapters/inbound/httpapi"
	"github.com/faber-numeris/sem/internal/infrastructure/config"
	infrapostgres "github.com/faber-numeris/sem/internal/infrastructure/postgres"
)

type App struct {
	server *http.Server
	pool   infrapostgres.AbstractPool
}

func NewAppWithDeps(ctx context.Context, cfg config.IAppConfig, connectFn ConnectFunc) (*App, error) {
	repo, pool, err := ProvideRepository(ctx, cfg, connectFn)
	if err != nil {
		return nil, err
	}

	svc := ProvideEmulatorService(repo, cfg.InstanceURL())
	router := httpapi.NewRouter(svc, repo.Ping)

	return &App{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port()),
			Handler: router,
		},
		pool: pool,
	}, nil
}

func NewApp(ctx context.Context) (*App, error) {
	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}
	return NewAppWithDeps(ctx, cfg, infrapostgres.Connect)
}

func (a *App) Run(ctx context.Context) error {
	slog.Info("Starting Salesforce Emulator service", "address", a.server.Addr)

	srvErrChan := make(chan error, 1)
	go func() {
		if srvErr := a.server.ListenAndServe(); srvErr != nil && !errors.Is(srvErr, http.ErrServerClosed) {
			srvErrChan <- srvErr
		}
	}()

	defer a.pool.Close()

	select {
	case err := <-srvErrChan:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		slog.Info("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return a.server.Shutdown(shutdownCtx)
	}
}

func Run(ctx context.Context) error {
	app, err := NewApp(ctx)
	if err != nil {
		return err
	}
	return app.Run(ctx)
}
