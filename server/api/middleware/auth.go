// Package middleware: Auth/JWT is quarantined — not a v1 product path (CONTEXT.md, PRD US28).
// Do not attach Auth() to Cast/Open/Stamp/Re-release/discovery routes.
package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ctxUserIDKey  = "userID"
	tokenDuration = 24 * time.Hour
)

func jwtSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is required")
	}
	return secret
}

// IssueToken — quarantined helper for non-product experiments only.
func IssueToken(userID int32) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret()))
}

// Auth — quarantined. Must not guard anonymous Visitor bottle flows.
func Auth() fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected signing method")
			}
			return []byte(jwtSecret()), nil
		})
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user_id in token")
		}

		c.Locals(ctxUserIDKey, int32(userIDFloat))

		return c.Next()
	}
}

// UserIDFromCtx reads quarantined auth locals; product handlers must not depend on this.
func UserIDFromCtx(c fiber.Ctx) (int32, bool) {
	id, ok := c.Locals(ctxUserIDKey).(int32)
	return id, ok
}
