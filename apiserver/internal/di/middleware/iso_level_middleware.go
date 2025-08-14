package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
)

const isoLevelKey = "di_isolation_level"

func withIsolationLevel(ctx context.Context, level pgx.TxIsoLevel) context.Context {
	return context.WithValue(ctx, isoLevelKey, level)
}

func GetIsolationLevel(ctx context.Context) pgx.TxIsoLevel {
	level := ctx.Value(isoLevelKey)
	if level == nil {
		var zero pgx.TxIsoLevel
		return zero
	}

	return level.(pgx.TxIsoLevel)
}

type IsoLevelMiddleware struct {
	logger *slog.Logger
}

func NewIsoLevelMiddleware(logger *slog.Logger) *IsoLevelMiddleware {
	return &IsoLevelMiddleware{
		logger: logger,
	}
}

func (m *IsoLevelMiddleware) Middleware(level pgx.TxIsoLevel, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := withIsolationLevel(r.Context(), level)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
