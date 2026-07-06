package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/model"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/server/http/api"
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

func (a *App) CreateEvent(ctx context.Context, event api.EventRequest) (api.EventResponse, error) {
	id := uuid.New().String()
	storageEvent := model.EventRequestToStorage(event, id)

	if err := a.storage.CreateEvent(ctx, storageEvent); err != nil {
		return api.EventResponse{}, err
	}

	return model.StorageToEventResponse(storageEvent), nil
}

func (a *App) UpdateEvent(ctx context.Context, id string, event api.EventRequest) (api.EventResponse, error) {
	storageEvent := model.EventRequestToStorage(event, id)

	if err := a.storage.UpdateEvent(ctx, storageEvent); err != nil {
		return api.EventResponse{}, err
	}

	return model.StorageToEventResponse(storageEvent), nil
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	return a.storage.DeleteEvent(ctx, id)
}

func (a *App) ListEventsDay(ctx context.Context, date time.Time) ([]api.EventResponse, error) {
	events, err := a.storage.ListEventsDay(ctx, date)
	if err != nil {
		return nil, err
	}
	return model.StorageListToEventResponse(events), nil
}

func (a *App) ListEventsWeek(ctx context.Context, date time.Time) ([]api.EventResponse, error) {
	events, err := a.storage.ListEventsWeek(ctx, date)
	if err != nil {
		return nil, err
	}
	return model.StorageListToEventResponse(events), nil
}

func (a *App) ListEventsMonth(ctx context.Context, date time.Time) ([]api.EventResponse, error) {
	events, err := a.storage.ListEventsMonth(ctx, date)
	if err != nil {
		return nil, err
	}
	return model.StorageListToEventResponse(events), nil
}
