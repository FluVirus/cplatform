package postgres

import (
	"context"
	"cplatform/internal/application/contracts/infrastructure"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWork struct {
	conn       *pgxpool.Conn
	txIsoLevel pgx.TxIsoLevel

	hasCurrTx bool
	currTx    pgx.Tx

	logger *slog.Logger

	userRepository *userRepository
}

func newUnitOfWorkWithIsoLevel(conn *pgxpool.Conn, logger *slog.Logger, txIsoLevel pgx.TxIsoLevel) *UnitOfWork {
	uow := &UnitOfWork{
		conn:       conn,
		logger:     logger,
		txIsoLevel: txIsoLevel,
	}

	return uow
}

func (uow *UnitOfWork) Tx(ctx context.Context) (pgx.Tx, error) {
	if !uow.hasCurrTx {
		tx, err := uow.conn.BeginTx(ctx, pgx.TxOptions{
			IsoLevel: uow.txIsoLevel,
		})

		if err != nil {
			return nil, err
		}

		uow.hasCurrTx = true
		uow.currTx = tx
	}

	return uow.currTx, nil
}

func (uow *UnitOfWork) SaveChanges(ctx context.Context) error {
	if !uow.hasCurrTx {
		return nil
	}

	err := uow.currTx.Commit(ctx)
	uow.hasCurrTx = false

	return err
}

func (uow *UnitOfWork) RollbackChanges(ctx context.Context) error {
	if !uow.hasCurrTx {
		return nil
	}

	err := uow.currTx.Rollback(ctx)
	uow.hasCurrTx = false

	return err
}

func (uow *UnitOfWork) UserRepository(context.Context) infrastructure.UserRepository {
	if uow.userRepository == nil {
		uow.userRepository = newUserRepository(uow, uow.logger)
	}

	return uow.userRepository
}

func (uow *UnitOfWork) Close(ctx context.Context) error {
	defer func() {
		uow.hasCurrTx = false
		uow.currTx = nil
		uow.conn.Release()
	}()

	if uow.hasCurrTx {
		newCtx := context.WithoutCancel(ctx)
		err := uow.currTx.Rollback(newCtx)
		if err != nil {
			return fmt.Errorf("failed to rollback transaction in uow_close: %w", err)
		}
	}

	return nil
}
