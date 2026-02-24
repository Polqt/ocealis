package handler

import (
	"time"

	"github.com/Polqt/ocealis/ws"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

var startTime = time.Now()

type HealthHandler struct {
	pool *pgxpool.Pool
	hub  *ws.Hub
}

func NewHealthHandler(pool *pgxpool.Pool, hub *ws.Hub) *HealthHandler {
	return &HealthHandler{
		pool: pool,
		hub:  hub,
	}
}

func (h *HealthHandler) Check(c fiber.Ctx) error {
	ctx := c.Context()
	dbStatus := "ok"
	if err := h.pool.Ping(ctx); err != nil {
		dbStatus = "unreachable"
	}

	status := fiber.StatusOK
	if dbStatus != "ok" {
		status = fiber.StatusServiceUnavailable
	}

	return c.Status(status).JSON(fiber.Map{
		"status":     dbStatus,
		"service":    "ocealis",
		"version":    "1.0.0",
		"uptime":     time.Since(startTime).String(),
		"ws_clients": h.hub.ClientCount(),
		"database":   dbStatus,
	})
}
