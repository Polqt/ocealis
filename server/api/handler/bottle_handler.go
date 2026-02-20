package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/Polqt/ocealis/api/middleware"
	"github.com/Polqt/ocealis/internal/service"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v3"
)

type createBottleRequest struct {
	MessageText string     `json:"message_text" validate:"required,min=1,max=1000"`
	BottleStyle int32      `json:"bottle_style"  validate:"min=0,max=9"`
	StartLat    float64    `json:"start_lat"     validate:"required,min=-90,max=90"`
	StartLng    float64    `json:"start_lng"     validate:"required,min=-180,max=180"`
	ReleaseAt   *time.Time `json:"release_at"`
}

type DiscoverBottleRequest struct {
	UserLat float64 `json:"user_lat" validate:"required,min=-90,max=90"`
	UserLng float64 `json:"user_lng" validate:"required,min=-180,max=180"`
}

type releaseBottleRequest struct {
	Lat float64 `json:"lat" validate:"required,min=-90,max=90"`
	Lng float64 `json:"lng" validate:"required,min=-180,max=180"`
}

type BottleHandler struct {
	svc      service.BottleService
	validate *validator.Validate
}

func NewBottleHandler(svc service.BottleService) *BottleHandler {
	return &BottleHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// Create a new bottle
func (h *BottleHandler) CreateBottle(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	var req createBottleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	bottle, err := h.svc.CreateBottle(c.Context(), service.CreateBottleInput{
		SenderID:    userID,
		MessageText: req.MessageText,
		BottleStyle: req.BottleStyle,
		StartLat:    req.StartLat,
		StartLng:    req.StartLng,
		ReleaseAt:   req.ReleaseAt,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not release bottle")
	}

	return c.Status(fiber.StatusCreated).JSON(bottle)
}

func (h *BottleHandler) GetBottle(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid bottle id")
	}

	bottle, err := h.svc.GetBottle(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "bottle not found")
	}

	return c.Status(fiber.StatusOK).JSON(bottle)
}

func (h *BottleHandler) GetJourney(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid bottle id")
	}

	journey, err := h.svc.GetJourney(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "journey not found")
	}

	return c.Status(fiber.StatusOK).JSON(journey)
}

func (h *BottleHandler) DiscoverBottle(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid bottle id")
	}

	var req DiscoverBottleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invald request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	journey, err := h.svc.DiscoverBottle(c.Context(), service.DiscoverBottleInput{
		BottleID:   id,
		DiscoverID: userID,
		UserLat:    req.UserLat,
		UserLng:    req.UserLng,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBottleNotFound):
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrAlreadyDiscovered):
			return fiber.NewError(fiber.StatusConflict, err.Error())
		case errors.Is(err, service.ErrSenderCannotDiscover):
			return fiber.NewError(fiber.StatusForbidden, err.Error())
		}
	}

	return c.Status(fiber.StatusOK).JSON(journey)

}

func (h *BottleHandler) ReleaseBottle(c fiber.Ctx) error {
	userID, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid bottle id")
	}

	var req releaseBottleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	bottle, err := h.svc.ReleaseBottle(c.Context(), id, userID, req.Lat, req.Lng)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not release bottle")
	}

	return c.Status(fiber.StatusOK).JSON(bottle)
}

func parseID(c fiber.Ctx, param string) (int32, error) {
	n, err := strconv.Atoi(c.Params(param))
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}
