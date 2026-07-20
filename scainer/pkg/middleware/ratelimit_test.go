package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/lksh/scainer/pkg/middleware"
)

func TestLoginRateLimit(t *testing.T) {
	e := echo.New()
	e.IPExtractor = middleware.ExtractIPFromProxyHeaders
	e.Use(middleware.LoginRateLimit(3, time.Minute))
	e.POST("/api/auth/login", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	e.GET("/api/contests", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	doLogin := func(xff string) int {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		if xff != "" {
			req.Header.Set(echo.HeaderXForwardedFor, xff)
		}
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	t.Run("allows under limit", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			if code := doLogin("10.0.0.1"); code != http.StatusNoContent {
				t.Fatalf("attempt %d: got %d want 204", i+1, code)
			}
		}
	})

	t.Run("blocks over limit", func(t *testing.T) {
		if code := doLogin("10.0.0.1"); code != http.StatusTooManyRequests {
			t.Fatalf("got %d want 429", code)
		}
	})

	t.Run("other IP independent", func(t *testing.T) {
		if code := doLogin("10.0.0.2"); code != http.StatusNoContent {
			t.Fatalf("got %d want 204", code)
		}
	})

	t.Run("non-login not limited", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			req := httptest.NewRequest(http.MethodGet, "/api/contests", nil)
			req.Header.Set(echo.HeaderXForwardedFor, "10.0.0.1")
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != http.StatusNoContent {
				t.Fatalf("got %d want 204", rec.Code)
			}
		}
	})
}
