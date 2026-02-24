package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// 60 requests per minute per IP
func RateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        60,
		Expiration: time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "slow down, the ocean is patient")
		},
	})
}

// 3 requests per minute per IP
func StrictRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        3,
		Expiration: time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "one bottle at a time")
		},
	})
}

// User Rate Limit is JWT-aware - keys by user_id for authenticated authenticated routes.
// Falls back to IP if no user_id is present (shouldn't happen on auth routes)
func UserRateLimit(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c fiber.Ctx) string {
			// User id from ctx reads from c.locals - only set after auth runs
			// this middleware must be registered after auth in the chain.
			if id, ok := UserIDFromCtx(c); ok {
				return fmt.Sprintf("user:%d", id)
			}
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "you are moving too fast")
		},
	})
}
