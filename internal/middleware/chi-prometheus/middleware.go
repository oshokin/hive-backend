package chi_prometheus

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

type Middleware struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

const (
	requestsTotalName   = "requests_total"
	requestDurationName = "response_time_seconds"
)

var (
	defaultBuckets     = []float64{0.001, 0.005, 0.015, 0.05, 0.1, 0.25, 0.5, 0.75, 1, 1.5, 2, 3.5, 5}
	httpRequestsLabels = []string{"method", "path", "status"}
)

func NewMiddleware(serviceName string, buckets ...float64) func(next http.Handler) http.Handler {
	if len(buckets) == 0 {
		buckets = defaultBuckets
	}

	m := Middleware{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        requestsTotalName,
				Help:        "Tracks the number of HTTP requests.",
				ConstLabels: prometheus.Labels{"service": serviceName},
			}, httpRequestsLabels),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        requestDurationName,
				Help:        "Tracks the latencies for HTTP requests in milliseconds.",
				ConstLabels: prometheus.Labels{"service": serviceName},
				Buckets:     buckets,
			}, httpRequestsLabels),
	}

	prometheus.MustRegister(
		m.requestsTotal,
		m.requestDuration)

	return m.handler
}

func (m *Middleware) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				ctx := chi.RouteContext(r.Context())
				routePattern := strings.Join(ctx.RoutePatterns, "")
				routePattern = strings.Replace(routePattern, "/*/", "/", -1)
				status := m.getStatusLabel(ww.Status())

				m.requestsTotal.WithLabelValues(
					r.Method,
					routePattern,
					status,
				).Inc()
				m.requestDuration.WithLabelValues(
					routePattern,
					r.Method,
					status).
					Observe(float64(time.Since(start).Seconds()))
			}()

			next.ServeHTTP(ww, r)
		})
}

func (m *Middleware) getStatusLabel(status int) string {
	switch {
	case status >= 500:
		return "error"
	case status == 404:
		return "not_found"
	case status >= 400:
		return "client_error"
	default:
		return "ok"
	}
}
