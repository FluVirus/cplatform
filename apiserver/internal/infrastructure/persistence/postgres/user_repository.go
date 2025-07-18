package postgres

import (
	"context"
	"cplatform/internal/application/contracts"
	"cplatform/internal/domain"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
)

type userRepository struct {
	logger *slog.Logger
	uow    *UnitOfWork
}

func newUserRepository(logger *slog.Logger) *userRepository {
	return &userRepository{
		logger: logger,
	}
}

func (repo *userRepository) setParentUnitOfWork(parent *UnitOfWork) {
	repo.uow = parent
}

func (repo *userRepository) AddUser(ctx context.Context, user *domain.User) error {
	tx, err := repo.uow.Tx()
	if err != nil {
		return fmt.Errorf("cannot fetch transaction: %w", err)
	}

	var id domain.UserId
	err = tx.QueryRow(ctx,
		"INSERT INTO public.users (name, email, password_hash, salt) VALUES ($1, $2, $3, $4) RETURNING id",
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Salt,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "c_unique_user_email" {
				err = fmt.Errorf("%w: %s", contracts.ErrDuplicateEmail, err.Error())
			}

			return fmt.Errorf("fail perform sql query: %w", err)
		}
	}

	user.Id = id

	return nil
}

func (repo *userRepository) DeleteUser(ctx context.Context, id domain.UserId) error {
	tx, err := repo.uow.Tx()
	if err != nil {
		return fmt.Errorf("cannot fetch transaction: %w", err)
	}

	newCtx := context.WithoutCancel(ctx)
	_, err = tx.Exec(newCtx, "DELETE FROM public.users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("fail perform sql query: %w", err)
	}

	return nil
}

func (repo *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	tx, err := repo.uow.Tx()
	if err != nil {
		return nil, fmt.Errorf("cannot fetch transaction: %w", err)
	}

	var userDto UserDto
	err = tx.QueryRow(ctx,
		"SELECT id, name, email, password_hash, salt FROM public.users WHERE email = $1",
		email,
	).Scan(&userDto)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: no user with such email", contracts.ErrUserNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("fail get user by email: %w", err)
	}

	var user = &domain.User{
		Id:           domain.UserId(userDto.Id),
		Name:         userDto.Name,
		Email:        userDto.Email,
		Salt:         userDto.Salt,
		PasswordHash: userDto.PasswordHash,
	}

	return user, nil
}
