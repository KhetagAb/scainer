package transport

import (
	"fmt"
	"log"
	"strings"

	"github.com/labstack/echo/v4"
)

const apiErrorLoggedKey = "scainer.api_error_logged"

func logAPIError(c echo.Context, status int, err error, fields ...any) {
	c.Set(apiErrorLoggedKey, true)

	var b strings.Builder
	fmt.Fprintf(&b, "api error method=%s path=%s status=%d", c.Request().Method, c.Request().URL.Path, status)
	if err != nil {
		fmt.Fprintf(&b, " err=%q", err.Error())
	}
	for i := 0; i+1 < len(fields); i += 2 {
		fmt.Fprintf(&b, " %v=%v", fields[i], fields[i+1])
	}
	log.Println(b.String())
}

func apiErrorFallbackMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			handlerErr := next(c)
			if c.Response().Status < 400 {
				return handlerErr
			}
			if logged, _ := c.Get(apiErrorLoggedKey).(bool); logged {
				return handlerErr
			}
			logAPIError(c, c.Response().Status, handlerErr)
			return handlerErr
		}
	}
}
