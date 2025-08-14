package middleware

import (
	"context"
	"cplatform/internal/di"
	"cplatform/internal/di/scope"
	"cplatform/pkg/slogext"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
)

const reqScopeKey = "di_request_scope"

func GetScope(ctx context.Context) *scope.Scope {
	s := ctx.Value(reqScopeKey)
	if s == nil {
		return nil
	}

	return s.(*scope.Scope)
}

func withScope(ctx context.Context, s *scope.Scope) context.Context {
	return context.WithValue(ctx, reqScopeKey, s)
}

type ScopeFactory interface {
	CreateScopeWithIsolationLevel(level pgx.TxIsoLevel) *scope.Scope
}

type ScopeMiddleware struct {
	logger          *slog.Logger
	reqScopeFactory ScopeFactory
}

func NewScopeMiddleware(logger *slog.Logger, reqScopeFactory ScopeFactory) *ScopeMiddleware {
	return &ScopeMiddleware{
		logger:          logger,
		reqScopeFactory: reqScopeFactory,
	}
}

func (m *ScopeMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		level := GetIsolationLevel(r.Context())
		if level == "" {
			level = di.DefaultIsoLevel
		}

		s := m.reqScopeFactory.CreateScopeWithIsolationLevel(level)
		ctx := withScope(r.Context(), s)

		defer func() {
			err := s.Close(ctx)
			if err != nil {
				m.logger.Warn("fail to close scope during user registration", slogext.Cause(err))
			}
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
