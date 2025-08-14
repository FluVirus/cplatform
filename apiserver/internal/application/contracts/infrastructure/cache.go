package infrastructure

import (
	"context"
	"cplatform/internal/domain"
)

type Cache interface {
	SaveUserByEmail(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}
