package postgres

import (
	"context"
	"cplatform/internal/application/contracts/infrastructure"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var defaultIsoLevel = pgx.ReadCommitted

type UnitOfWorkFactory struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewUnitOfWorkFactory(pool *pgxpool.Pool, logger *slog.Logger) *UnitOfWorkFactory {
	return &UnitOfWorkFactory{
		pool:   pool,
		logger: logger,
	}
}

func (f *UnitOfWorkFactory) Create(ctx context.Context) infrastructure.UnitOfWork {
	return f.CreateWithIsolationLevel(ctx, defaultIsoLevel)
}

func (f *UnitOfWorkFactory) CreateWithIsolationLevel(ctx context.Context, txIsoLevel pgx.TxIsoLevel) infrastructure.UnitOfWork {
	conn, err := f.pool.Acquire(ctx)
	if err != nil {
		err = fmt.Errorf("failed to create unit of work: %w", err)
		panic(err)
	}

	return newUnitOfWorkWithIsoLevel(conn, f.logger, txIsoLevel)
}
