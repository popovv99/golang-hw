package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/server/http/api"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// EventRequestToStorage конвертирует EventRequest в storage.Event
func EventRequestToStorage(req api.EventRequest, id string) storage.Event {
	event := storage.Event{
		ID:           id,
		Title:        req.Title,
		Date:         req.Date,
		EndDate:      req.EndDate,
		UserID:       req.UserId.String(),
		NotifyBefore: 0,
	}

	if req.Description != nil {
		event.Description = *req.Description
	}

	if req.NotifyBefore != nil {
		// Используем Go duration format (например, "24h", "30m", "1h30m")
		// вместо ISO 8601 для простоты и совместимости с time.ParseDuration
		duration, err := time.ParseDuration(*req.NotifyBefore)
		if err == nil {
			event.NotifyBefore = duration
		}
	}

	return event
}

// StorageToEventResponse конвертирует storage.Event в EventResponse
func StorageToEventResponse(event storage.Event) api.EventResponse {
	parsedUUID := types.UUID(uuid.MustParse(event.ID))
	parsedUserID := types.UUID(uuid.MustParse(event.UserID))
	response := api.EventResponse{
		Id:           &parsedUUID,
		Title:        &event.Title,
		Date:         &event.Date,
		EndDate:      &event.EndDate,
		Description:  &event.Description,
		UserId:       &parsedUserID,
		NotifyBefore: nil,
	}

	if event.NotifyBefore > 0 {
		duration := event.NotifyBefore.String()
		response.NotifyBefore = &duration
	}

	return response
}

// StorageListToEventResponse конвертирует []storage.Event в []EventResponse
func StorageListToEventResponse(events []storage.Event) []api.EventResponse {
	result := make([]api.EventResponse, len(events))
	for i, event := range events {
		result[i] = StorageToEventResponse(event)
	}
	return result
}
