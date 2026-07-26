package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"scainer/pkg/auth"
)

var registry = prometheus.NewRegistry()

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scainer_http_requests_total",
			Help: "Total HTTP requests by method, Echo route template and status code.",
		},
		[]string{"method", "route", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "scainer_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)
	teacherRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scainer_teacher_requests_total",
			Help: "Authenticated HTTP requests per teacher login and Echo route template.",
		},
		[]string{"username", "route"},
	)
	authLoginsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scainer_auth_logins_total",
			Help: "Login attempts by outcome (success or failed).",
		},
		[]string{"status"},
	)
)

func init() {
	registry.MustRegister(httpRequestsTotal, httpRequestDuration, teacherRequestsTotal, authLoginsTotal)
}

func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

func ObserveAuthLogin(success bool) {
	status := "failed"
	if success {
		status = "success"
	}
	authLoginsTotal.WithLabelValues(status).Inc()
}

// HTTPMiddleware records HTTP and per-teacher request metrics after the handler runs.
// Place after JWT middleware so username is available in request context.
func HTTPMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/metrics" {
				return next(c)
			}

			start := time.Now()
			err := next(c)

			route := c.Path()
			if route == "" {
				route = c.Request().URL.Path
			}
			method := c.Request().Method
			status := strconv.Itoa(c.Response().Status)
			elapsed := time.Since(start).Seconds()

			httpRequestsTotal.WithLabelValues(method, route, status).Inc()
			httpRequestDuration.WithLabelValues(method, route).Observe(elapsed)

			if login, ok := auth.LoginFrom(c.Request().Context()); ok {
				teacherRequestsTotal.WithLabelValues(login, route).Inc()
			}

			return err
		}
	}
}
