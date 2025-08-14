package scope

import (
	"context"
	"cplatform/internal/application/contracts/application"
	"cplatform/internal/application/contracts/infrastructure"
	"cplatform/internal/application/users"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
)

type Scope struct {
	factory  *Factory
	isoLevel pgx.TxIsoLevel

	uow    infrastructure.UnitOfWork
	uowMux sync.Mutex

	userServiceMux sync.Mutex
	userService    application.UserService
}

func (s *Scope) UserService(ctx context.Context) application.UserService {
	if s.userService == nil {
		s.userServiceMux.Lock()
		defer s.userServiceMux.Unlock()

		if s.userService == nil {
			s.userService = users.NewUserService(s.UnitOfWork(ctx), s.factory.cache, s.factory.logger)
		}
	}

	return s.userService
}

func (s *Scope) UnitOfWork(ctx context.Context) infrastructure.UnitOfWork {
	if s.uow == nil {
		s.uowMux.Lock()
		defer s.uowMux.Unlock()

		s.uow = s.factory.uowFactory.CreateWithIsolationLevel(ctx, s.isoLevel)
	}

	return s.uow
}

func (s *Scope) Close(ctx context.Context) error {
	err := s.uow.Close(ctx)
	if err != nil {
		return fmt.Errorf("fail close req scoped services: %w", err)
	}

	return nil
}
