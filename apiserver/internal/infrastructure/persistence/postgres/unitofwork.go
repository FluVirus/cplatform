package postgres

import (
	"context"
	"cplatform/internal/application/contracts"
	"cplatform/pkg/slogext"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
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
	userRepo := newUserRepository(logger)

	uow := &UnitOfWork{
		conn:           conn,
		logger:         logger,
		txIsoLevel:     txIsoLevel,
		userRepository: userRepo,
	}

	userRepo.setParentUnitOfWork(uow)

	return uow
}

func (uow *UnitOfWork) Tx() (pgx.Tx, error) {
	if !uow.hasCurrTx {
		tx, err := uow.conn.BeginTx(context.TODO(), pgx.TxOptions{
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

func (uow *UnitOfWork) UserRepository() contracts.UserRepository {
	return uow.userRepository
}

func (uow *UnitOfWork) Close() error {
	if uow.hasCurrTx {
		uow.hasCurrTx = false

		err := uow.currTx.Rollback(context.Background())
		if err != nil {
			uow.logger.Warn("failed to rollback transaction in uow_close", slogext.Cause(err))
		}

		uow.currTx = nil
	}

	uow.conn.Release()

	return nil
}
