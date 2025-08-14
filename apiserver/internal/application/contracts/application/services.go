package application

import (
	"context"
	"cplatform/internal/domain"
)

type UserService interface {
	RegisterUser(ctx context.Context, name string, email string, password string) error
	GetUserWithCheckCredentials(ctx context.Context, email, password string) (*domain.User, error)
	DeleteUser(ctx context.Context, id domain.UserId) error
}
