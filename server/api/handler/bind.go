package handler

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

// bindJSON decodes the request body with encoding/json.
// Fiber v3 Bind().JSON also runs binder validation that conflicts with
// go-playground tags and commonly surfaces as opaque 400s.
func bindJSON(c fiber.Ctx, dst any) error {
	body := c.Body()
	if len(body) == 0 {
		return fmt.Errorf("empty body")
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}
