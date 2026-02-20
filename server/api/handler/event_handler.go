package handler

import (
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

	events, err := h.events.GetByBottleID(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve events")
	}

	return c.JSON(fiber.Map{
		"bottle_id": id,
		"events":    events,
		"count":     len(events),
	})
}
