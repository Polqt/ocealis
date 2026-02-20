package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/Polqt/ocealis/util"
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

// Issue Token generates a JWT token for an anonymous user and returns it.
// Called once when a user first visits the site, and the token is stored in localStorage for subsequent requests.
func IssueToken(userID int32) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret()))
}

// Auth Validates the bearer token on every protected route and extracts the user ID from the token claims, making it available in the request context for downstream handlers.
// Injects userID into c.Locals for handlers to access.
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

		// JWT numbers are deserialized as float64, so we need to convert it to int32
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user_id in token")
		}

		c.Locals(ctxUserIDKey, int32(userIDFloat))

		return c.Next()
	}
}

// This function extracts the authenticated user ID from the request context, which was set by the Auth middleware. It returns the user ID and a boolean indicating whether the user is authenticated.
func UserIDFromCtx(c fiber.Ctx) (int32, bool) {
	id, ok := c.Locals(ctxUserIDKey).(int32)
	return id, ok
}
