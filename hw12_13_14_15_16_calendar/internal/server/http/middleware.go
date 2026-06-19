package internalhttp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
)

func loggingMiddleware(next http.Handler, eventLogger logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter для перехвата статуса
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		latency := time.Since(start)

		// Формируем лог в формате: IP [date] method path version status latency "user-agent"
		ip := r.RemoteAddr
		date := time.Now().Format("02/Jan/2006:15:04:05 -0700")
		method := r.Method
		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		version := r.Proto
		status := rw.statusCode
		userAgent := r.UserAgent()

		eventLogger.Info(fmt.Sprintf("%s [%s] %s %s %s %d %d \"%s\"",
			ip, date, method, path, version, status, latency.Milliseconds(), userAgent))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
