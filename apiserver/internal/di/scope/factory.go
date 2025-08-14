package scope

import (
	"cplatform/internal/application/contracts/infrastructure"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type Factory struct {
	uowFactory infrastructure.UnitOfWorkFactory
	cache      infrastructure.Cache
	logger     *slog.Logger
}

func NewFactory(
	uowFactory infrastructure.UnitOfWorkFactory,
	cache infrastructure.Cache,
	logger *slog.Logger,
) *Factory {
	return &Factory{
		uowFactory: uowFactory,
		cache:      cache,
		logger:     logger,
	}
}

func (f *Factory) CreateScopeWithIsolationLevel(txLevel pgx.TxIsoLevel) *Scope {
	return &Scope{
		factory:  f,
		isoLevel: txLevel,
	}
}
