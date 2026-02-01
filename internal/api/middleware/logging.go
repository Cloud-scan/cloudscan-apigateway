package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// RequestLogger logs HTTP requests
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			// Log request
			log.WithFields(log.Fields{
				"method":      c.Request().Method,
				"path":        c.Request().URL.Path,
				"status":      c.Response().Status,
				"duration_ms": time.Since(start).Milliseconds(),
				"ip":          c.RealIP(),
				"user_agent":  c.Request().UserAgent(),
			}).Info("HTTP request")

			return err
		}
	}
}