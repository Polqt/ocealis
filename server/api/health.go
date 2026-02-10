package api

import (
	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/gofiber/fiber/v3"
)

type HealthHandler struct {
	queries *ocealis.Queries
}

func NewHealthHandler(q *ocealis.Queries) *HealthHandler {
	return &HealthHandler{
		queries: q,
	}
}

func (h *HealthHandler) HealthCheck(c fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "Ocealis API",
		"version": "1.0.0",
	})
}

func RegisterHealthRoutes(router fiber.Router, handler *HealthHandler) {
	router.Get("/health", handler.HealthCheck)
}
