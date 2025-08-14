package basic

import (
	"context"
	"cplatform/internal/application/contracts/application"
	"cplatform/internal/di/middleware"
	"cplatform/internal/domain"
	"cplatform/pkg/slogext"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

const userKey = "basic_user"

type BasicAuthMiddleware struct {
	logger *slog.Logger
}

func NewBasicAuthMiddleware(logger *slog.Logger) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{
		logger: logger,
	}
}

// TODO: migrate to coded api errors

func (m *BasicAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		authHeaderParts := strings.SplitN(authHeader, " ", 2)
		if len(authHeaderParts) < 2 && strings.ToLower(authHeaderParts[0]) != "basic" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(authHeaderParts[1])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		payloadParts := strings.SplitN(string(payload), ":", 2)
		if len(payloadParts) != 2 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		email := payloadParts[0]
		password := payloadParts[1]

		scope := middleware.GetScope(r.Context())
		if scope == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user, err := scope.UserService(r.Context()).GetUserWithCheckCredentials(r.Context(), email, password)
		if err != nil {
			m.logger.Error("fail to check user during basic auth", slogext.Cause(err))

			if errors.Is(err, application.ErrUserNotFound) ||
				errors.Is(err, application.ErrWrongCredentials) {
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			return
		}

		ctx := withUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(reqCtx context.Context) *domain.User {
	user := reqCtx.Value(userKey)
	if user == nil {
		return nil
	}

	return user.(*domain.User)
}

func withUser(reqCtx context.Context, user *domain.User) context.Context {
	return context.WithValue(reqCtx, userKey, user)
}
