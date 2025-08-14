package redis

import (
	"context"
	"cplatform/internal/domain"
	"errors"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	logger *slog.Logger
}

func NewRedisCache(client *redis.Client, logger *slog.Logger) *Cache {
	return &Cache{
		client: client,
		logger: logger,
	}
}

func (r *Cache) SaveUserByEmail(ctx context.Context, user *domain.User) error {
	dto := UserDto{
		Id:           int64(user.Id),
		Name:         user.Name,
		PasswordHash: user.PasswordHash,
		Salt:         user.Salt,
	}

	err := r.client.Set(ctx, user.Email, dto, 0).Err()
	if err != nil {
		return fmt.Errorf("could not save user: %w", err)
	}

	return nil
}

func (r *Cache) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var dto UserDto
	err := r.client.Get(ctx, email).Scan(&dto)
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("could not get user by email: %w", err)
	}

	user := &domain.User{
		Id:           domain.UserId(dto.Id),
		Name:         dto.Name,
		Email:        email,
		PasswordHash: dto.PasswordHash,
		Salt:         dto.Salt,
	}

	return user, nil
}
