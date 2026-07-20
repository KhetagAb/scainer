package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/lksh/scainer/pkg/auth"
	"github.com/lksh/scainer/pkg/middleware"
)

func TestRequireJWT(t *testing.T) {
	svc := auth.New("admin", "secret", "jwt-secret", time.Hour)
	token, _, err := svc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	e.Use(middleware.RequireJWT(svc))
	e.POST("/api/auth/login", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	e.GET("/api/contests", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	t.Run("login without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status: got %d want %d body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
		}
	})

	t.Run("contests without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/contests", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d want 401", rec.Code)
		}
	})

	t.Run("contests invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/contests", nil)
		req.Header.Set("Authorization", "Bearer not-valid")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status: got %d want 401", rec.Code)
		}
	})

	t.Run("contests valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/contests", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d want 200 body=%s", rec.Code, rec.Body.String())
		}
		if rec.Body.String() != "ok" {
			t.Fatalf("body: got %q", rec.Body.String())
		}
	})
}
