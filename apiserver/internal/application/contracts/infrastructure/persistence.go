package infrastructure

import (
	"context"
	"cplatform/internal/domain"
	"errors"

	"github.com/jackc/pgx/v5"
)

var (
	ErrDuplicateEmail = errors.New("email already exists")
	ErrUserNotFound   = errors.New("user not found")
)

type UserRepository interface {
	AddUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id domain.UserId) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type UnitOfWork interface {
	UserRepository(ctx context.Context) UserRepository
	SaveChanges(ctx context.Context) error
	RollbackChanges(ctx context.Context) error
	Close(ctx context.Context) error
}

type UnitOfWorkFactory interface {
	Create(ctx context.Context) UnitOfWork

	// CreateWithIsolationLevel has the note: txIsolationLevel is tradeoff not to write huge isolation level detection logic
	CreateWithIsolationLevel(ctx context.Context, level pgx.TxIsoLevel) UnitOfWork
}
