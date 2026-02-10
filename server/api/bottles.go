package api

import (
	"strconv"

	"github.com/Polqt/ocealis/services"
	"github.com/gofiber/fiber/v3"
)

type BottleHandler struct {
	bottleService *services.BottleService
}

func NewBottleHandler(bs *services.BottleService) *BottleHandler {
	return &BottleHandler{
		bottleService: bs,
	}
}

func (h *BottleHandler) GetBottleById(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid bottle id",
		})
	}

	bottle, err := h.bottleService.GetBottleById(c.Context(), int32(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "bottle not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(bottle)
}


func RegisterBottleRoutes(router fiber.Router, handler *BottleHandler) {
	bottles := router.Group("/bottles")
	
	bottles.Get("/:id", handler.GetBottleById)
}