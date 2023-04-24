package chi_prometheus

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

// Middleware is a struct that represents a middleware that collects metrics
// about HTTP requests and their duration.
// It contains a CounterVec to track the number of requests
// and a HistogramVec to track the duration of requests in milliseconds.
// The middleware registers these metrics with Prometheus.
type Middleware struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

const (
	httpLabel           = "http"
	requestsTotalName   = "requests_total"
	requestDurationName = "response_time_seconds"

	statusClientErr = 400
	statusNotFound  = 404
	statusError     = 500
)

var (
	defaultBuckets     = []float64{0.001, 0.005, 0.015, 0.05, 0.1, 0.25, 0.5, 0.75, 1, 1.5, 2, 3.5, 5}
	httpRequestsLabels = []string{"type", "method", "path", "status"}
)

// NewMiddleware creates a new Prometheus middleware handler that provides
// request counting and duration metrics for a Go-Chi HTTP server.
func NewMiddleware(serviceName string, buckets ...float64) func(next http.Handler) http.Handler {
	if len(buckets) == 0 {
		buckets = defaultBuckets
	}

	m := Middleware{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        requestsTotalName,
				Help:        "Number of requests.",
				ConstLabels: prometheus.Labels{"service": serviceName},
			}, httpRequestsLabels),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:        requestDurationName,
				Help:        "Latencies for requests in milliseconds.",
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
				routePattern := m.getRoutePattern(r)
				status := m.getStatusLabel(ww.Status())

				m.requestsTotal.WithLabelValues(
					httpLabel,
					r.Method,
					routePattern,
					status,
				).Inc()
				m.requestDuration.WithLabelValues(
					httpLabel,
					r.Method,
					routePattern,
					status).
					Observe(time.Since(start).Seconds())
			}()

			next.ServeHTTP(ww, r)
		})
}

func (m *Middleware) getRoutePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if pattern := rctx.RoutePattern(); pattern != "" {
		return pattern
	}

	routePath := r.URL.Path
	if r.URL.RawPath != "" {
		routePath = r.URL.RawPath
	}

	tctx := chi.NewRouteContext()
	if !rctx.Routes.Match(tctx, r.Method, routePath) {
		return routePath
	}

	return tctx.RoutePattern()
}

func (m *Middleware) getStatusLabel(status int) string {
	switch {
	case status >= statusError:
		return "error"
	case status == statusNotFound:
		return "not_found"
	case status >= statusClientErr:
		return "client_error"
	default:
		return "ok"
	}
}
