package api

import (
	"strings"

	"github.com/Polqt/ocealis/api/handler"
	"github.com/Polqt/ocealis/api/middleware"
	"github.com/Polqt/ocealis/ws"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type Handlers struct {
	Health *handler.HealthHandler
	Bottle *handler.BottleHandler
	User   *handler.UserHandler
	Event  *handler.EventHandler
}

// RegisterRoutes wires all HTTP and WebSocket routes onto app.
func RegisterRoutes(app *fiber.App, h Handlers, hub *ws.Hub, log *zap.Logger) {
	app.Get("/api/health", h.Health.Check)

	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Ocealis API",
			"version": "1.0.0",
		})
	})

	// WebSocket — upgrade check then stream drift events to connected clients.
	app.Use("/ws", func(c fiber.Ctx) error {
		if strings.EqualFold(c.Get("Upgrade"), "websocket") {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", ws.NewDriftHandler(hub, log))

	v1 := app.Group("/api/v1")

	// User routes
	users := v1.Group("/users")
	users.Post("/", middleware.StrictRateLimit(), h.User.CreateUser)
	users.Get("/profile", middleware.Auth(), h.User.GetUser)

	// Bottle routes
	bottles := v1.Group("/bottles", middleware.Auth())
	bottles.Post("/", middleware.StrictRateLimit(), h.Bottle.CreateBottle)
	bottles.Get("/:id", h.Bottle.GetBottle)
	bottles.Get("/:id/journey", h.Bottle.GetJourney)
	bottles.Get("/:id/events", h.Event.GetBottleEvents)
	bottles.Post("/:id/discover", h.Bottle.DiscoverBottle)
	bottles.Post("/:id/release", h.Bottle.ReleaseBottle)
}
