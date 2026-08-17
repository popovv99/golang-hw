package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once           sync.Once
	defaultMetrics *Metrics
)

// Default returns the package-level metrics instance.
func Default() *Metrics {
	once.Do(func() {
		defaultMetrics = New()
	})
	return defaultMetrics
}

// New creates a new Metrics instance.
func New() *Metrics {
	m := &Metrics{
		httpRequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "calendar_http_requests_total",
			Help: "Total number of HTTP requests handled by the calendar API",
		}, []string{"method", "path", "status"}),

		httpRequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "calendar_http_request_duration_seconds",
			Help:    "HTTP request latency distribution in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"method", "path"}),

		eventsCreated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "calendar_events_created_total",
			Help: "Total number of events created",
		}),

		eventsUpdated: promauto.NewCounter(prometheus.CounterOpts{
			Name: "calendar_events_updated_total",
			Help: "Total number of events updated",
		}),

		eventsDeleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "calendar_events_deleted_total",
			Help: "Total number of events deleted",
		}),

		notificationsSent: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "calendar_notifications_sent_total",
			Help: "Total number of notifications sent by the scheduler",
		}, []string{"status"}),

		notificationsSaved: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "calendar_notifications_saved_total",
			Help: "Total number of notifications saved by the storer",
		}, []string{"status"}),

		schedulerRuns: promauto.NewCounter(prometheus.CounterOpts{
			Name: "calendar_scheduler_runs_total",
			Help: "Total number of scheduler iterations executed",
		}),

		schedulerErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "calendar_scheduler_errors_total",
			Help: "Total number of failed scheduler iterations",
		}),

		schedulerLastSuccess: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "calendar_scheduler_last_success_timestamp_seconds",
			Help: "Unix timestamp of the last successful scheduler iteration",
		}),

		storerRuns: promauto.NewCounter(prometheus.CounterOpts{
			Name: "calendar_storer_runs_total",
			Help: "Total number of messages processed by the storer",
		}),

		storerErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "calendar_storer_errors_total",
			Help: "Total number of storer processing errors",
		}),

		storerLastSuccess: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "calendar_storer_last_success_timestamp_seconds",
			Help: "Unix timestamp of the last successful storer message processing",
		}),
	}
	return m
}

// Metrics holds all Prometheus collectors for the calendar service.
type Metrics struct {
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	eventsCreated        prometheus.Counter
	eventsUpdated        prometheus.Counter
	eventsDeleted        prometheus.Counter
	notificationsSent    *prometheus.CounterVec
	notificationsSaved   *prometheus.CounterVec
	schedulerRuns        prometheus.Counter
	schedulerErrors      prometheus.Counter
	schedulerLastSuccess prometheus.Gauge
	storerRuns           prometheus.Counter
	storerErrors         prometheus.Counter
	storerLastSuccess    prometheus.Gauge
}

// RecordHTTP records an HTTP request and its duration.
func (m *Metrics) RecordHTTP(method, path string, statusCode int, duration time.Duration) {
	m.httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// IncEventsCreated increments the created events counter.
func (m *Metrics) IncEventsCreated() {
	m.eventsCreated.Inc()
}

// IncEventsUpdated increments the updated events counter.
func (m *Metrics) IncEventsUpdated() {
	m.eventsUpdated.Inc()
}

// IncEventsDeleted increments the deleted events counter.
func (m *Metrics) IncEventsDeleted() {
	m.eventsDeleted.Inc()
}

// IncNotificationsSent increments the sent notifications counter.
func (m *Metrics) IncNotificationsSent(status string) {
	m.notificationsSent.WithLabelValues(status).Inc()
}

// IncNotificationsSaved increments the saved notifications counter.
func (m *Metrics) IncNotificationsSaved(status string) {
	m.notificationsSaved.WithLabelValues(status).Inc()
}

// IncSchedulerRuns increments the scheduler runs counter.
func (m *Metrics) IncSchedulerRuns() {
	m.schedulerRuns.Inc()
}

// IncSchedulerErrors increments the scheduler errors counter.
func (m *Metrics) IncSchedulerErrors() {
	m.schedulerErrors.Inc()
}

// SetSchedulerLastSuccess sets the last successful scheduler iteration timestamp.
func (m *Metrics) SetSchedulerLastSuccess(t time.Time) {
	m.schedulerLastSuccess.Set(float64(t.Unix()))
}

// IncStorerRuns increments the storer runs counter.
func (m *Metrics) IncStorerRuns() {
	m.storerRuns.Inc()
}

// IncStorerErrors increments the storer errors counter.
func (m *Metrics) IncStorerErrors() {
	m.storerErrors.Inc()
}

// SetStorerLastSuccess sets the last successful storer message processing timestamp.
func (m *Metrics) SetStorerLastSuccess(t time.Time) {
	m.storerLastSuccess.Set(float64(t.Unix()))
}
