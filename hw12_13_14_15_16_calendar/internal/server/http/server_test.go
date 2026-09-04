package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/server/http/api"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// mockApp - мок для Application интерфейса
type mockApp struct {
	events map[string]storage.Event
}

func newMockApp() *mockApp {
	return &mockApp{
		events: make(map[string]storage.Event),
	}
}

func (m *mockApp) CreateEvent(ctx context.Context, event api.EventRequest) (api.EventResponse, error) {
	id := uuid.New()
	storageEvent := storage.Event{
		ID:           id.String(),
		Title:        event.Title,
		Date:         event.Date,
		EndDate:      event.EndDate,
		UserID:       event.UserId.String(),
		NotifyBefore: 0,
	}
	if event.Description != nil {
		storageEvent.Description = *event.Description
	}
	m.events[id.String()] = storageEvent
	uuidPtr := types.UUID(id)
	return api.EventResponse{
		Id:      &uuidPtr,
		Title:   &event.Title,
		Date:    &event.Date,
		EndDate: &event.EndDate,
		UserId:  &event.UserId,
	}, nil
}

func (m *mockApp) UpdateEvent(ctx context.Context, id string, event api.EventRequest) (api.EventResponse, error) {
	if _, exists := m.events[id]; !exists {
		return api.EventResponse{}, storage.ErrEventNotFound
	}
	storageEvent := storage.Event{
		ID:           id,
		Title:        event.Title,
		Date:         event.Date,
		EndDate:      event.EndDate,
		UserID:       event.UserId.String(),
		NotifyBefore: 0,
	}
	if event.Description != nil {
		storageEvent.Description = *event.Description
	}
	m.events[id] = storageEvent
	parsedUUID := uuid.MustParse(id)
	uuidPtr := types.UUID(parsedUUID)
	return api.EventResponse{
		Id:      &uuidPtr,
		Title:   &event.Title,
		Date:    &event.Date,
		EndDate: &event.EndDate,
		UserId:  &event.UserId,
	}, nil
}

func (m *mockApp) DeleteEvent(ctx context.Context, id string) error {
	if _, exists := m.events[id]; !exists {
		return storage.ErrEventNotFound
	}
	delete(m.events, id)
	return nil
}

func (m *mockApp) ListEventsDay(ctx context.Context, date time.Time) ([]api.EventResponse, error) {
	var result []api.EventResponse
	for _, event := range m.events {
		if event.Date.Year() == date.Year() && event.Date.Month() == date.Month() && event.Date.Day() == date.Day() {
			parsedUUID := uuid.MustParse(event.ID)
			uuidPtr := types.UUID(parsedUUID)
			parsedUserID := uuid.MustParse(event.UserID)
			userIDPtr := types.UUID(parsedUserID)
			result = append(result, api.EventResponse{
				Id:      &uuidPtr,
				Title:   &event.Title,
				Date:    &event.Date,
				EndDate: &event.EndDate,
				UserId:  &userIDPtr,
			})
		}
	}
	return result, nil
}

func (m *mockApp) ListEventsWeek(ctx context.Context, date time.Time) ([]api.EventResponse, error) {
	var result []api.EventResponse
	for _, event := range m.events {
		// Простая проверка: событие в пределах 7 дней от даты
		diff := event.Date.Sub(date)
		if diff >= 0 && diff < 7*24*time.Hour {
			parsedUUID := uuid.MustParse(event.ID)
			uuidPtr := types.UUID(parsedUUID)
			parsedUserID := uuid.MustParse(event.UserID)
			userIDPtr := types.UUID(parsedUserID)
			result = append(result, api.EventResponse{
				Id:      &uuidPtr,
				Title:   &event.Title,
				Date:    &event.Date,
				EndDate: &event.EndDate,
				UserId:  &userIDPtr,
			})
		}
	}
	return result, nil
}

func (m *mockApp) ListEventsMonth(ctx context.Context, date time.Time) ([]api.EventResponse, error) {
	var result []api.EventResponse
	monthStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	for _, event := range m.events {
		if (event.Date.Equal(monthStart) || event.Date.After(monthStart)) && event.Date.Before(monthEnd) {
			parsedUUID := uuid.MustParse(event.ID)
			uuidPtr := types.UUID(parsedUUID)
			parsedUserID := uuid.MustParse(event.UserID)
			userIDPtr := types.UUID(parsedUserID)
			result = append(result, api.EventResponse{
				Id:      &uuidPtr,
				Title:   &event.Title,
				Date:    &event.Date,
				EndDate: &event.EndDate,
				UserId:  &userIDPtr,
			})
		}
	}
	return result, nil
}

func TestCreateEvent(t *testing.T) {
	app := newMockApp()
	h := &handler{app: app, logger: nil}

	userID := uuid.New()
	reqBody := api.EventRequest{
		Title:   "Test Event",
		Date:    time.Now().Add(1 * time.Hour),
		EndDate: time.Now().Add(2 * time.Hour),
		UserId:  types.UUID(userID),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response api.EventResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Title == nil || *response.Title != "Test Event" {
		t.Errorf("expected title 'Test Event', got %v", response.Title)
	}
}

func TestCreateEventInvalidBody(t *testing.T) {
	app := newMockApp()
	h := &handler{app: app, logger: nil}

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateEvent(t *testing.T) {
	app := newMockApp()
	h := &handler{app: app, logger: nil}

	// Сначала создаем событие
	id := uuid.New()
	userID := uuid.New()
	event := storage.Event{
		ID:      id.String(),
		Title:   "Original Title",
		Date:    time.Now().Add(1 * time.Hour),
		EndDate: time.Now().Add(2 * time.Hour),
		UserID:  userID.String(),
	}
	app.events[id.String()] = event

	// Обновляем событие
	reqBody := api.EventRequest{
		Title:   "Updated Title",
		Date:    time.Now().Add(1 * time.Hour),
		EndDate: time.Now().Add(2 * time.Hour),
		UserId:  types.UUID(userID),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/events/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req, types.UUID(id))

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response api.EventResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Title == nil || *response.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %v", response.Title)
	}
}

func TestUpdateEventNotFound(t *testing.T) {
	app := newMockApp()
	h := &handler{app: app, logger: nil}

	id := uuid.New()
	userID := uuid.New()
	reqBody := api.EventRequest{
		Title:   "Updated Title",
		Date:    time.Now().Add(1 * time.Hour),
		EndDate: time.Now().Add(2 * time.Hour),
		UserId:  types.UUID(userID),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/events/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req, types.UUID(id))

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteEvent(t *testing.T) {
	app := newMockApp()
	h := &handler{app: app, logger: nil}

	id := uuid.New()
	userID := uuid.New()
	event := storage.Event{
		ID:      id.String(),
		Title:   "Test Event",
		Date:    time.Now().Add(1 * time.Hour),
		EndDate: time.Now().Add(2 * time.Hour),
		UserID:  userID.String(),
	}
	app.events[id.String()] = event

	req := httptest.NewRequest(http.MethodDelete, "/events/"+id.String(), nil)
	w := httptest.NewRecorder()

	h.DeleteEvent(w, req, types.UUID(id))

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	if _, exists := app.events[id.String()]; exists {
		t.Error("event should be deleted")
	}
}

func TestDeleteEventNotFound(t *testing.T) {
	app := newMockApp()
	h := &handler{app: app, logger: nil}

	id := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/events/"+id.String(), nil)
	w := httptest.NewRecorder()

	h.DeleteEvent(w, req, types.UUID(id))

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestListEventsDay(t *testing.T) {
	app := newMockApp()
	h := &handler{app: app, logger: nil}

	now := time.Now()
	userID := uuid.New()
	event := storage.Event{
		ID:      uuid.New().String(),
		Title:   "Test Event",
		Date:    now,
		EndDate: now.Add(1 * time.Hour),
		UserID:  userID.String(),
	}
	app.events[event.ID] = event

	params := api.ListEventsDayParams{Date: now}
	req := httptest.NewRequest(http.MethodGet, "/events/day?date="+url.QueryEscape(now.Format(time.RFC3339)), nil)
	w := httptest.NewRecorder()

	h.ListEventsDay(w, req, params)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []api.EventResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 event, got %d", len(response))
	}
}

func TestListEventsWeek(t *testing.T) {
	app := newMockApp()
	h := &handler{app: app, logger: nil}

	now := time.Now()
	userID := uuid.New()
	event := storage.Event{
		ID:      uuid.New().String(),
		Title:   "Test Event",
		Date:    now.Add(24 * time.Hour),
		EndDate: now.Add(25 * time.Hour),
		UserID:  userID.String(),
	}
	app.events[event.ID] = event

	params := api.ListEventsWeekParams{Date: now}
	req := httptest.NewRequest(http.MethodGet, "/events/week?date="+url.QueryEscape(now.Format(time.RFC3339)), nil)
	w := httptest.NewRecorder()

	h.ListEventsWeek(w, req, params)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []api.EventResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 event, got %d", len(response))
	}
}

func TestListEventsMonth(t *testing.T) {
	app := newMockApp()
	h := &handler{app: app, logger: nil}

	now := time.Now()
	userID := uuid.New()
	event := storage.Event{
		ID:      uuid.New().String(),
		Title:   "Test Event",
		Date:    now.Add(5 * 24 * time.Hour),
		EndDate: now.Add(6 * 24 * time.Hour),
		UserID:  userID.String(),
	}
	app.events[event.ID] = event

	params := api.ListEventsMonthParams{Date: now}
	req := httptest.NewRequest(http.MethodGet, "/events/month?date="+url.QueryEscape(now.Format(time.RFC3339)), nil)
	w := httptest.NewRecorder()

	h.ListEventsMonth(w, req, params)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response []api.EventResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 event, got %d", len(response))
	}
}
