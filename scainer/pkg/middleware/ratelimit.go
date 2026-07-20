package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

func LoginRateLimit(max int, window time.Duration) echo.MiddlewareFunc {
	if max <= 0 {
		max = 5
	}
	if window <= 0 {
		window = time.Minute
	}
	lim := &loginLimiter{max: max, window: window, hits: make(map[string][]time.Time)}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method != http.MethodPost || c.Path() != "/api/auth/login" {
				return next(c)
			}
			ip := c.RealIP()
			if ip == "" {
				ip = c.Request().RemoteAddr
			}
			if !lim.allow(ip) {
				return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "too many login attempts"})
			}
			return next(c)
		}
	}
}

type loginLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	hits   map[string][]time.Time
}

func (l *loginLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	prev := l.hits[key]
	recent := make([]time.Time, 0, len(prev)+1)
	for _, t := range prev {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= l.max {
		l.hits[key] = recent
		return false
	}
	l.hits[key] = append(recent, now)
	return true
}
