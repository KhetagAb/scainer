package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func LoginRateLimit(max int, window time.Duration) echo.MiddlewareFunc {
	if max <= 0 {
		max = 5
	}
	if window <= 0 {
		window = time.Minute
	}
	return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().Method != http.MethodPost || c.Path() != "/api/auth/login"
		},
		Store: echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{
			Rate:      rate.Every(window / time.Duration(max)),
			Burst:     max,
			ExpiresIn: window,
		}),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			ip := c.RealIP()
			if ip == "" {
				return c.Request().RemoteAddr, nil
			}
			return ip, nil
		},
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "too many login attempts"})
		},
	})
}
