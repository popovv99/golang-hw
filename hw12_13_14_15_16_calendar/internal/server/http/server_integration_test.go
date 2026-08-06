package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/server/http/api"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// setupTestServer создаёт тестовый HTTP-сервер с реальным роутером
func setupTestServer(app *mockApp) *httptest.Server {
	mux := chi.NewRouter()
	handler := &handler{app: app, logger: nil}
	apiHandler := api.HandlerFromMux(handler, mux)
	return httptest.NewServer(apiHandler)
}

func TestIntegrationCreateEvent(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	userID := uuid.New()
	reqBody := api.EventRequest{
		Title:   "Test Event",
		Date:    time.Now().Add(1 * time.Hour),
		EndDate: time.Now().Add(2 * time.Hour),
		UserId:  types.UUID(userID),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var response api.EventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Title == nil || *response.Title != "Test Event" {
		t.Errorf("expected title 'Test Event', got %v", response.Title)
	}
}

func TestIntegrationCreateEventInvalidBody(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/events", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestIntegrationUpdateEvent(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	// Сначала создаем событие напрямую через mock
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

	// Обновляем через HTTP
	reqBody := api.EventRequest{
		Title:   "Updated Title",
		Date:    time.Now().Add(1 * time.Hour),
		EndDate: time.Now().Add(2 * time.Hour),
		UserId:  types.UUID(userID),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPut, server.URL+"/events/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response api.EventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Title == nil || *response.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %v", response.Title)
	}
}

func TestIntegrationUpdateEventInvalidUUID(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	userID := uuid.New()
	reqBody := api.EventRequest{
		Title:   "Updated Title",
		Date:    time.Now().Add(1 * time.Hour),
		EndDate: time.Now().Add(2 * time.Hour),
		UserId:  types.UUID(userID),
	}
	body, _ := json.Marshal(reqBody)

	// Передаём невалидный UUID
	req, _ := http.NewRequest(http.MethodPut, server.URL+"/events/invalid-uuid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// oapi-codegen должен вернуть 400 для невалидного UUID
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid UUID, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestIntegrationUpdateEventNotFound(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	id := uuid.New()
	userID := uuid.New()
	reqBody := api.EventRequest{
		Title:   "Updated Title",
		Date:    time.Now().Add(1 * time.Hour),
		EndDate: time.Now().Add(2 * time.Hour),
		UserId:  types.UUID(userID),
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPut, server.URL+"/events/"+id.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestIntegrationDeleteEvent(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

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

	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/events/"+id.String(), nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}

	if _, exists := app.events[id.String()]; exists {
		t.Error("event should be deleted")
	}
}

func TestIntegrationDeleteEventInvalidUUID(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/events/invalid-uuid", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid UUID, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestIntegrationDeleteEventNotFound(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	id := uuid.New()
	req, _ := http.NewRequest(http.MethodDelete, server.URL+"/events/"+id.String(), nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestIntegrationListEventsDay(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

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

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/events/day?date="+url.QueryEscape(now.Format(time.RFC3339)), nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response []api.EventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 event, got %d", len(response))
	}
}

func TestIntegrationListEventsDayInvalidDate(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	// Передаём невалидную дату
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/events/day?date=invalid-date", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// oapi-codegen должен вернуть 400 для невалидной даты
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid date, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestIntegrationListEventsWeek(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

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

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/events/week?date="+url.QueryEscape(now.Format(time.RFC3339)), nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response []api.EventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 event, got %d", len(response))
	}
}

func TestIntegrationListEventsWeekInvalidDate(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/events/week?date=invalid-date", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid date, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestIntegrationListEventsMonth(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

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

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/events/month?date="+url.QueryEscape(now.Format(time.RFC3339)), nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response []api.EventResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response) != 1 {
		t.Errorf("expected 1 event, got %d", len(response))
	}
}

func TestIntegrationListEventsMonthInvalidDate(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/events/month?date=invalid-date", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid date, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestIntegrationRouteNotFound(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/invalid-route", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d for invalid route, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestIntegrationMethodNotAllowed(t *testing.T) {
	app := newMockApp()
	server := setupTestServer(app)
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/events/day", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d for method not allowed, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}
