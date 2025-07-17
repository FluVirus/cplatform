package postgres

import (
	"context"
	"cplatform/internal/application/contracts"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

var defaultIsoLevel = pgx.ReadCommitted

type Factory struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewUnitOfWorkFactory(pool *pgxpool.Pool, logger *slog.Logger) *Factory {
	return &Factory{
		pool:   pool,
		logger: logger,
	}
}

func (f *Factory) Create(ctx context.Context) (contracts.UnitOfWork, error) {
	return f.CreateWithIsolationLevel(ctx, defaultIsoLevel)
}

func (f *Factory) CreateWithIsolationLevel(ctx context.Context, txIsoLevel pgx.TxIsoLevel) (contracts.UnitOfWork, error) {
	conn, err := f.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create unit of work: %w", err)
	}

	return newUnitOfWorkWithIsoLevel(conn, f.logger, txIsoLevel), nil
}
