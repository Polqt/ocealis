package handler

import (
	"errors"
	"strconv"

	"github.com/Polqt/ocealis/api/middleware"
	"github.com/Polqt/ocealis/internal/cast"
	"github.com/Polqt/ocealis/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type createBottleRequest struct {
	Nickname       string   `json:"nickname" validate:"required,min=1,max=24"`
	MessageText    string   `json:"message_text" validate:"required,min=1,max=500"`
	BottleStyle    int32    `json:"bottle_style" validate:"min=0,max=9"`
	StartLat       *float64 `json:"start_lat" validate:"omitempty,min=-90,max=90"`
	StartLng       *float64 `json:"start_lng" validate:"omitempty,min=-180,max=180"`
	TurnstileToken string   `json:"turnstile_token" validate:"required"`
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
	turnstile middleware.TurnstileVerifier
	validate *validator.Validate
}

func NewBottleHandler(svc service.BottleService, turnstile middleware.TurnstileVerifier) *BottleHandler {
	if turnstile == nil {
		turnstile = middleware.AcceptTurnstile{}
	}
	return &BottleHandler{
		svc:       svc,
		turnstile: turnstile,
		validate:  validator.New(),
	}
}

// CreateBottle is Cast — anonymous Visitor, no JWT, no caster tracking.
func (h *BottleHandler) CreateBottle(c fiber.Ctx) error {
	var req createBottleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := h.turnstile.Verify(c.Context(), req.TurnstileToken, c.IP()); err != nil {
		return fiber.NewError(fiber.StatusForbidden, "cast blocked")
	}

	// one of lat/lng missing → treat as no geo (basin fallback in service)
	var lat, lng *float64
	if req.StartLat != nil && req.StartLng != nil {
		lat, lng = req.StartLat, req.StartLng
	}

	bottle, err := h.svc.CreateBottle(c.Context(), service.CreateBottleInput{
		Nickname:    req.Nickname,
		MessageText: req.MessageText,
		BottleStyle: req.BottleStyle,
		StartLat:    lat,
		StartLng:    lng,
	})
	if err != nil {
		switch {
		case errors.Is(err, cast.ErrNicknameRequired),
			errors.Is(err, cast.ErrNicknameTooLong),
			errors.Is(err, cast.ErrMessageRequired),
			errors.Is(err, cast.ErrMessageTooLong):
			return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
		default:
			return fiber.NewError(fiber.StatusInternalServerError, "could not cast bottle")
		}
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

// DiscoverBottle is legacy claim path — Open must not claim (issue 03). No JWT.
func (h *BottleHandler) DiscoverBottle(c fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid bottle id")
	}

	var req DiscoverBottleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	journey, err := h.svc.DiscoverBottle(c.Context(), service.DiscoverBottleInput{
		BottleID:   id,
		DiscoverID: 0, // anonymous Visitor
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
		default:
			return fiber.NewError(fiber.StatusInternalServerError, "could not open bottle")
		}
	}

	return c.Status(fiber.StatusOK).JSON(journey)
}

// ReleaseBottle is Re-release — anonymous, Nickname in later issue.
func (h *BottleHandler) ReleaseBottle(c fiber.Ctx) error {
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

	bottle, err := h.svc.ReleaseBottle(c.Context(), id, 0, req.Lat, req.Lng)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not re-release bottle")
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
