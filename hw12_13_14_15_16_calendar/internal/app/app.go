package app

import (
	"context"

	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

type App struct {
	logger  logger.Logger
	storage storage.Storage
}

func New(eventLogger logger.Logger, eventStorage storage.Storage) *App {
	return &App{
		logger:  eventLogger,
		storage: eventStorage,
	}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	// TODO: реализовать логику создания события
	return nil
}

// TODO
