package controller

import (
	"cplatform/internal/application/authentication/basic"
	"cplatform/internal/di/middleware"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/gorilla/mux"
)

type Router struct {
	controller   *Controller
	useIsoLevel  *middleware.IsoLevelMiddleware
	useScope     *middleware.ScopeMiddleware
	useBasicAuth *basic.BasicAuthMiddleware
	logger       *slog.Logger
}

func NewRouter(
	controller *Controller,
	useIsoLevel *middleware.IsoLevelMiddleware,
	useScope *middleware.ScopeMiddleware,
	useBasicAuth *basic.BasicAuthMiddleware,
	logger *slog.Logger,
) *Router {
	return &Router{
		controller:   controller,
		useIsoLevel:  useIsoLevel,
		useScope:     useScope,
		useBasicAuth: useBasicAuth,
		logger:       logger,
	}
}

func (r *Router) CreateHandler() http.Handler {
	m := mux.NewRouter()

	api := m.PathPrefix("/api").Subrouter()

	v1 := api.PathPrefix("/v1").Subrouter()

	v1.Handle("/users",
		r.useIsoLevel.Middleware(pgx.ReadCommitted,
			r.useScope.Middleware(
				http.HandlerFunc(r.controller.RegisterUserHandler)))).
		Methods(http.MethodPost)

	v1.Handle("/users",
		r.useIsoLevel.Middleware(pgx.ReadCommitted,
			r.useScope.Middleware(
				r.useBasicAuth.Middleware(
					http.HandlerFunc(r.controller.DeleteSelfUserHandler))))).
		Methods(http.MethodDelete)

	return m
}
