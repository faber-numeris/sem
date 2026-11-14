package bootstrap

import (
	"context"

	outboundpg "github.com/faber-numeris/sem/internal/adapters/outbound/postgres"
	"github.com/faber-numeris/sem/internal/core/services"
	"github.com/faber-numeris/sem/internal/infrastructure/config"
	infrapostgres "github.com/faber-numeris/sem/internal/infrastructure/postgres"
	"github.com/faber-numeris/sem/internal/ports/inbound"
	"github.com/faber-numeris/sem/internal/ports/outbound"
)

type ConnectFunc func(ctx context.Context, cfg config.IDatabaseConfig) (infrapostgres.AbstractPool, error)

func ProvideRepository(ctx context.Context, cfg config.IDatabaseConfig, connectFn ConnectFunc) (outbound.Repository, infrapostgres.AbstractPool, error) {
	pool, err := connectFn(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	repo := outboundpg.NewRepository(pool)
	return repo, pool, nil
}

func ProvideEmulatorService(repo outbound.Repository, defaultInstanceURL string) inbound.EmulatorService {
	return services.NewService(repo, defaultInstanceURL)
}
