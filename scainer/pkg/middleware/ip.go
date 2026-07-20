package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ExtractIPFromProxyHeaders: первый X-Forwarded-For, иначе X-Real-IP, иначе RemoteAddr.
// Scainer в проде доступен только из docker-сети nginx — заголовкам прокси доверяем.
func ExtractIPFromProxyHeaders(req *http.Request) string {
	if xff := req.Header.Get(echo.HeaderXForwardedFor); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if rip := req.Header.Get(echo.HeaderXRealIP); rip != "" {
		return strings.TrimSpace(rip)
	}
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}
