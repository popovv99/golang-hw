package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/metrics"
)

func loggingMiddleware(eventLogger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

		// Для метрик используем route-паттерн, чтобы не раздувать cardinality идентификаторами.
		// chi устанавливает r.Pattern в routeHTTP() для каждого совпавшего маршрута.
		// Для несовпавших (404/405) r.Pattern пуст — используем raw path.
		metricPath := r.Pattern
		if metricPath == "" {
			metricPath = r.URL.Path
		}
		// Не записываем /metrics, чтобы Prometheus-скрейп не засорял HTTP-метрики
		if r.URL.Path != "/metrics" {
			metrics.Default().RecordHTTP(r.Method, metricPath, rw.statusCode, latency)
		}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
