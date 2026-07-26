package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"scainer/pkg/auth"
)

// Кроме точного POST /api/auth/login.
func RequireJWT(svc auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodGet && c.Request().URL.Path == "/metrics" {
				return next(c)
			}
			if c.Request().Method == http.MethodPost && c.Path() == "/api/auth/login" {
				return next(c)
			}

			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			claims, err := svc.ParseToken(header)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			c.Set(auth.ContextUsernameKey, claims.Username)
			req := c.Request().WithContext(auth.WithLogin(c.Request().Context(), claims.Username))
			c.SetRequest(req)
			return next(c)
		}
	}
}
