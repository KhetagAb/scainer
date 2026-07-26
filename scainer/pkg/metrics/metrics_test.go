package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"scainer/pkg/auth"
	"scainer/pkg/metrics"
)

func TestHTTPMiddleware_recordsTeacherAndHTTPMetrics(t *testing.T) {
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request().WithContext(auth.WithLogin(c.Request().Context(), "alice"))
			c.SetRequest(req)
			return next(c)
		}
	})
	e.Use(metrics.HTTPMiddleware())
	e.GET("/api/contests", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/metrics", echo.WrapHandler(metrics.Handler()))

	req := httptest.NewRequest(http.MethodGet, "/api/contests", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	e.ServeHTTP(metricsRec, metricsReq)
	if metricsRec.Code != http.StatusOK {
		t.Fatalf("metrics status: got %d", metricsRec.Code)
	}

	body := metricsRec.Body.String()
	if !strings.Contains(body, `scainer_http_requests_total{method="GET",route="/api/contests",status="200"} 1`) {
		t.Fatalf("missing http counter in metrics:\n%s", body)
	}
	if !strings.Contains(body, `scainer_teacher_requests_total{route="/api/contests",username="alice"} 1`) {
		t.Fatalf("missing teacher counter in metrics:\n%s", body)
	}
	if strings.Contains(body, `scainer_teacher_requests_total{route="/metrics"`) {
		t.Fatalf("metrics scrape should not be counted as teacher activity:\n%s", body)
	}
}

func TestObserveAuthLogin(t *testing.T) {
	metrics.ObserveAuthLogin(true)
	metrics.ObserveAuthLogin(false)

	e := echo.New()
	e.GET("/metrics", echo.WrapHandler(metrics.Handler()))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `scainer_auth_logins_total{status="success"}`) {
		t.Fatalf("missing success login counter:\n%s", body)
	}
	if !strings.Contains(body, `scainer_auth_logins_total{status="failed"}`) {
		t.Fatalf("missing failed login counter:\n%s", body)
	}
}
