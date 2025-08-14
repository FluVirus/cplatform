package pipeline

import (
	"cplatform/internal/presentation/http/controller"
	"log/slog"
	"net/http"

	"github.com/rs/cors"
)

type Pipeline struct {
	router     *controller.Router
	useRecover *RecoverMiddleware
	useCors    *cors.Cors
	logger     *slog.Logger
}

func NewPipeline(router *controller.Router, useRecover *RecoverMiddleware, useCors *cors.Cors, logger *slog.Logger) *Pipeline {
	return &Pipeline{
		router:     router,
		useRecover: useRecover,
		useCors:    useCors,
		logger:     logger,
	}
}

func (p *Pipeline) CreateHandler() http.Handler {
	handler := p.router.CreateHandler()
	handler = p.useCors.Handler(handler)
	handler = p.useRecover.Middleware(handler)

	return handler
}
