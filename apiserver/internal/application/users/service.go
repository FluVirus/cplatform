package users

import (
	"bytes"
	"context"
	"cplatform/internal/application/contracts"
	"cplatform/internal/domain"
	"cplatform/pkg/slogext"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"log/slog"
	"math/rand/v2"
)

var (
	ErrDuplicateEmail   = errors.New("email already exists")
	ErrUserNotFound     = errors.New("user not found")
	ErrWrongCredentials = errors.New("wrong credentials")
)

type UserService struct {
	uow        contracts.UnitOfWork
	cache      contracts.Cache
	logger     *slog.Logger
	saltLength int
}

func NewUserService(uow contracts.UnitOfWork, cache contracts.Cache, logger *slog.Logger, saltLength int) *UserService {
	return &UserService{
		uow:        uow,
		cache:      cache,
		logger:     logger,
		saltLength: saltLength,
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

	err := s.uow.UserRepository().AddUser(ctx, user)

	if err != nil {
		if errors.Is(err, contracts.ErrDuplicateEmail) {
			err = fmt.Errorf("%w: %s", ErrDuplicateEmail, err.Error())
		}

		rollbackErr := s.uow.RollbackChanges(context.WithoutCancel(ctx))
		err = errors.Join(err, rollbackErr)
		
		return fmt.Errorf("fail user registration: %w", err)

	}

	err = s.uow.SaveChanges(ctx)
	if err != nil {
		return fmt.Errorf("fail saving changes: %w", err)
	}

	return nil
}

func (s *UserService) GetUserWithCheckCredentials(ctx context.Context, email, password string) (*domain.User, error) {
	var user *domain.User

	user, err := s.cache.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		if err != nil {
			s.logger.Warn("error during fetching user from redis", slogext.Cause(err))
		}

		user, err = s.uow.UserRepository().GetUserByEmail(ctx, email)

		if err != nil {
			if errors.Is(err, contracts.ErrUserNotFound) {
				err = fmt.Errorf("%w: %s", ErrUserNotFound, err.Error())
			}

			return nil, fmt.Errorf("fail get user: %w", err)
		}

		err = s.cache.SaveUserByEmail(ctx, user)
		if err != nil {
			s.logger.Warn("error during saving user to cache", slogext.Cause(err))
		}
	}

	hash := hashFunc([]byte(password), user.Salt)

	if !bytes.Equal(user.PasswordHash, hash) {
		return nil, ErrWrongCredentials
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
