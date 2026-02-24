package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Request Logger returns a fiber middleware that logs every HTTP request
// as a structured log with the method, path, status code, latency, and error (if any).
func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// Let the request pass through all subsequent handlers
		err := c.Next()

		// After the response is written, log the outcome
		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Choose log level based on status code
		// 5xx = error, 4xx = warning, else info
		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("latency", duration),
			zap.String("ip", c.IP()),
		}

		switch {
		case status >= 500:
			log.Error("request", fields...)
		case status >= 400:
			log.Warn("request", fields...)
		default:
			log.Info("request", fields...)
		}

		return err
	}
}
