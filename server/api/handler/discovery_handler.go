package handler

import (
	"strconv"

	"github.com/Polqt/ocealis/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type findNearbyRequest struct {
	Lat      *float64 `query:"lat"       validate:"required,min=-90,max=90"`
	Lng      *float64 `query:"lng"       validate:"required,min=-180,max=180"`
	RadiusKm float64  `query:"radius_km" validate:"omitempty,min=1,max=2000"`
	Limit    int32    `query:"limit"     validate:"omitempty,min=1,max=100"`
	Cursor   *int32   `query:"cursor"`
}

type browseMapRequest struct {
	MinLat *float64 `query:"min_lat" validate:"required,min=-90,max=90"`
	MaxLat *float64 `query:"max_lat" validate:"required,min=-90,max=90"`
	MinLng *float64 `query:"min_lng" validate:"required,min=-180,max=180"`
	MaxLng *float64 `query:"max_lng" validate:"required,min=-180,max=180"`
	Zoom   *float64 `query:"zoom"    validate:"required,min=0,max=22"`
}

type DiscoveryHandler struct {
	svc      service.DiscoveryService
	validate *validator.Validate
}

func NewDiscoveryHandler(svc service.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{svc: svc, validate: validator.New()}
}

// BrowseMap handles GET /discovery/map — heat vs Corks by zoom/viewport.
func (h *DiscoveryHandler) BrowseMap(c fiber.Ctx) error {
	var req browseMapRequest
	if err := c.Bind().Query(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	}
	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	if *req.MinLat > *req.MaxLat || *req.MinLng > *req.MaxLng {
		return fiber.NewError(fiber.StatusBadRequest, "viewport bounds inverted")
	}

	result, err := h.svc.BrowseMap(c.Context(), service.BrowseMapInput{
		MinLat: *req.MinLat,
		MaxLat: *req.MaxLat,
		MinLng: *req.MinLng,
		MaxLng: *req.MaxLng,
		Zoom:   *req.Zoom,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "map query failed")
	}
	return c.Status(fiber.StatusOK).JSON(result)
}

// FindNearby handles GET /discovery/ to find discoverable bottles near a location.
func (h *DiscoveryHandler) FindNearby(c fiber.Ctx) error {
	var req findNearbyRequest
	if err := c.Bind().Query(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid query parameters")
	}
	if err := h.validate.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	var cursor *int32
	if raw := c.Query("cursor"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid cursor")
		}
		v := int32(n)
		cursor = &v
	}

	result, err := h.svc.FindNearby(c.Context(), service.FindNearbyInput{
		Lat:      *req.Lat,
		Lng:      *req.Lng,
		RadiusKm: req.RadiusKm,
		Limit:    req.Limit,
		Cursor:   cursor,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "discovery failed")
	}

	return c.Status(fiber.StatusOK).JSON(result)
}
