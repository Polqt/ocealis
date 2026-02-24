package handler

import (
	"strconv"

	"github.com/Polqt/ocealis/internal/repository"
	"github.com/gofiber/fiber/v3"
)

type EventHandler struct {
	events repository.EventRepository
}

func NewEventHandler(events repository.EventRepository) *EventHandler {
	return &EventHandler{
		events: events,
	}
}

func (h *EventHandler) GetBottleEvents(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid bottle id")
	}

	// Parse optional cursor from query string: ?cursor=42
	var cursor *int32
	if raw := c.Query("cursor"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid cursor")
		}
		v := int32(n)
		cursor = &v
	}

	// Parse limit, default 20, max 100
	limit := int32(20)
	if raw := c.Query("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			return fiber.NewError(fiber.StatusBadRequest, "limit must be between 1 and 100")
		}
		limit = int32(n)
	}

	result, err := h.events.GetPaginated(c.Context(), repository.GetEventParams{
		BottleID: id,
		Cursor:   cursor,
		Limit:    limit,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve events")
	}

	return c.JSON(result)
}
