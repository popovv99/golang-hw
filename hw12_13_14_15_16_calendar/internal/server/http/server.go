package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/runtime/types"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/server/http/api"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

type Server struct {
	server *http.Server
	logger logger.Logger
	app    Application
}

type Application interface {
	CreateEvent(ctx context.Context, event api.EventRequest) (api.EventResponse, error)
	UpdateEvent(ctx context.Context, id string, event api.EventRequest) (api.EventResponse, error)
	DeleteEvent(ctx context.Context, id string) error
	ListEventsDay(ctx context.Context, date time.Time) ([]api.EventResponse, error)
	ListEventsWeek(ctx context.Context, date time.Time) ([]api.EventResponse, error)
	ListEventsMonth(ctx context.Context, date time.Time) ([]api.EventResponse, error)
}

func NewServer(eventLogger logger.Logger, app Application) *Server {
	return &Server{
		logger: eventLogger,
		app:    app,
	}
}

func (s *Server) Start(ctx context.Context, host, port string) error {
	mux := chi.NewRouter()

	// Регистрируем сгенерированные хендлеры
	handler := &handler{
		app:    s.app,
		logger: s.logger,
	}

	apiHandler := api.HandlerFromMux(handler, mux)

	// Оборачиваем в middleware для логирования
	loggingHandler := loggingMiddleware(apiHandler, s.logger)

	addr := net.JoinHostPort(host, port)
	s.logger.Info(fmt.Sprintf("server is starting on %s", addr))

	s.server = &http.Server{
		Addr:    addr,
		Handler: loggingHandler,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handler реализует интерфейс api.ServerInterface
type handler struct {
	app    Application
	logger logger.Logger
}

func (h *handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req api.EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	event, err := h.app.CreateEvent(r.Context(), req)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func (h *handler) UpdateEvent(w http.ResponseWriter, r *http.Request, id types.UUID) {
	var req api.EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", "BAD_REQUEST")
		return
	}

	event, err := h.app.UpdateEvent(r.Context(), id.String(), req)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(event)
}

func (h *handler) DeleteEvent(w http.ResponseWriter, r *http.Request, id types.UUID) {
	err := h.app.DeleteEvent(r.Context(), id.String())
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) ListEventsDay(w http.ResponseWriter, r *http.Request, params api.ListEventsDayParams) {
	events, err := h.app.ListEventsDay(r.Context(), params.Date)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}

func (h *handler) ListEventsWeek(w http.ResponseWriter, r *http.Request, params api.ListEventsWeekParams) {
	events, err := h.app.ListEventsWeek(r.Context(), params.Date)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}

func (h *handler) ListEventsMonth(w http.ResponseWriter, r *http.Request, params api.ListEventsMonthParams) {
	events, err := h.app.ListEventsMonth(r.Context(), params.Date)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(events)
}

func strPtr(s string) *string {
	return &s
}

func writeErrorResponse(w http.ResponseWriter, status int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(api.ErrorResponse{
		Message: strPtr(message),
		Code:    strPtr(code),
	})
}

func handleError(w http.ResponseWriter, err error, logger logger.Logger) {
	if errors.Is(err, storage.ErrDateBusy) {
		writeErrorResponse(w, http.StatusBadRequest, "date is busy", "DATE_BUSY")
		return
	}
	if errors.Is(err, storage.ErrEventNotFound) {
		writeErrorResponse(w, http.StatusNotFound, "event not found", "EVENT_NOT_FOUND")
		return
	}
	logger.Error("internal error: " + err.Error())
	writeErrorResponse(w, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
}
