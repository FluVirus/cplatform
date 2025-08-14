package users

import (
	"bytes"
	"context"
	"cplatform/internal/application/contracts/application"
	"cplatform/internal/application/contracts/infrastructure"
	"cplatform/internal/domain"
	"cplatform/pkg/slogext"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"

	"golang.org/x/crypto/argon2"
)

type UserService struct {
	uow        infrastructure.UnitOfWork
	cache      infrastructure.Cache
	logger     *slog.Logger
	saltLength int
}

func NewUserService(uow infrastructure.UnitOfWork, cache infrastructure.Cache, logger *slog.Logger) *UserService {
	return &UserService{
		uow:        uow,
		cache:      cache,
		logger:     logger,
		saltLength: 10,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, name string, email string, password string) error {
	salt := make([]byte, s.saltLength)

	for i := range len(salt) {
		salt[i] = byte(rand.IntN(256))
	}

	hash := hashFunc([]byte(password), salt)
	user := &domain.User{
		Name:         name,
		Email:        email,
		Salt:         salt,
		PasswordHash: hash,
	}

	err := s.uow.UserRepository(ctx).AddUser(ctx, user)

	if err != nil {
		if errors.Is(err, infrastructure.ErrDuplicateEmail) {
			err = fmt.Errorf("%w: %s", application.ErrDuplicateEmail, err.Error())
		}

		rollbackErr := s.uow.RollbackChanges(context.WithoutCancel(ctx))
		err = errors.Join(err, rollbackErr)

		return fmt.Errorf("fail user registration: %w", err)
	}

	return nil
}

func (s *UserService) GetUserWithCheckCredentials(ctx context.Context, email, password string) (*domain.User, error) {
	var user *domain.User

	user, err := s.cache.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("fail fetch user from redis", slogext.Cause(err))
	}

	if err != nil || user == nil {
		user, err = s.uow.UserRepository(ctx).GetUserByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, infrastructure.ErrUserNotFound) {
				err = fmt.Errorf("%w: %s", application.ErrUserNotFound, err.Error())
			}
			return nil, fmt.Errorf("fail get user: %w", err)
		}

		if cacheErr := s.cache.SaveUserByEmail(ctx, user); cacheErr != nil {
			s.logger.Warn("fail save user to cache", slogext.Cause(cacheErr))
		}
	}

	hash := hashFunc([]byte(password), user.Salt)
	if !bytes.Equal(user.PasswordHash, hash) {
		return nil, application.ErrWrongCredentials
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id domain.UserId) error {
	return errors.New("not implemented")
}

func hashFunc(password []byte, salt []byte) []byte {
	return argon2.IDKey(
		password,
		salt,
		1,
		64*1024,
		4,
		32,
	)
}
