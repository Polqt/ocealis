package handler

import (
	"strconv"

	"github.com/Polqt/ocealis/internal/repository"
	"github.com/gofiber/fiber/v3"
)

type OceanHandler struct {
	bottles repository.BottleRepository
}

func NewOceanHandler(bottles repository.BottleRepository) *OceanHandler {
	return &OceanHandler{bottles: bottles}
}

// ListBottles handles GET /api/v1/ocean/bottles — ambient map seed data.
func (h *OceanHandler) ListBottles(c fiber.Ctx) error {
	limit := int32(100)
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 500 {
			return fiber.NewError(fiber.StatusBadRequest, "limit must be between 1 and 500")
		}
		limit = int32(n)
	}

	bottles, err := h.bottles.ListOcean(c.Context(), limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list ocean bottles")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": bottles,
	})
}
