package basic

import (
	"context"
	"cplatform/internal/application/users"
	"cplatform/internal/domain"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// TODO: make as middleware

type Auth struct {
	userService users.UserService
	logger      *slog.Logger
}

func NewBasicAuth(userService users.UserService, logger *slog.Logger) *Auth {
	return &Auth{
		userService: userService,
		logger:      logger,
	}
}

func (m *Auth) Authenticate(w http.ResponseWriter, r *http.Request) (*domain.User, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, nil
	}

	authHeaderParts := strings.SplitN(authHeader, " ", 2)
	if authHeaderParts[0] != "Basic" {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, nil
	}

	payload, err := base64.StdEncoding.DecodeString(authHeaderParts[1])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, nil
	}

	payloadParts := strings.SplitN(string(payload), ":", 2)
	if len(payloadParts) != 2 {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, nil
	}

	email := payloadParts[0]
	password := payloadParts[1]

	user, err := m.userService.GetUserWithCheckCredentials(r.Context(), email, password)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) ||
			errors.Is(err, users.ErrWrongCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
		}

		// TODO: add coded api error
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	return user, nil
}

func GetUser(reqCtx context.Context) *domain.User {
	return reqCtx.Value("user").(*domain.User)
}
