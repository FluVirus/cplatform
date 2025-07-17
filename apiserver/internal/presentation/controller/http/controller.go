package http

import (
	"cplatform/internal/application/contracts"
	"log/slog"
)

type Controller struct {
	// TODO: create DI in another way
	uowFactory contracts.UnitOfWorkFactory
	cache      contracts.Cache
	logger     *slog.Logger
}

func NewController(
	uowFactory contracts.UnitOfWorkFactory,
	cache contracts.Cache,
	logger *slog.Logger,
) *Controller {
	return &Controller{
		uowFactory: uowFactory,
		cache:      cache,
		logger:     logger,
	}
}
