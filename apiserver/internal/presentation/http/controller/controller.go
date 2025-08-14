package controller

import (
	"log/slog"
)

type Controller struct {
	logger *slog.Logger
}

func NewController(logger *slog.Logger) *Controller {
	return &Controller{
		logger: logger,
	}
}
