package internalhttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
)

type Server struct {
	server *http.Server
	logger logger.Logger
	app    Application
}

type Application interface {
	// TODO: добавить методы приложения
}

func NewServer(eventLogger logger.Logger, app Application) *Server {
	return &Server{
		logger: eventLogger,
		app:    app,
	}
}

func (s *Server) Start(ctx context.Context, host, port string) error {
	mux := http.NewServeMux()
	mux.Handle("/", loggingMiddleware(http.HandlerFunc(s.helloHandler), s.logger))

	s.server = &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: mux,
	}

	s.logger.Info(fmt.Sprintf("server is starting on %s:%s", host, port))

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

func (s *Server) helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hello, World!")); err != nil {
		s.logger.Error("failed to write response: " + err.Error())
	}
}
