package api

import (
	"github.com/Polqt/ocealis/api/handler"
	"github.com/Polqt/ocealis/api/middleware"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type Handlers struct {
	Health *handler.HealthHandler
	Bottle *handler.BottleHandler
	User   *handler.UserHandler
	Event  *handler.EventHandler
}

func RegisterRoutes(app *fiber.App, h Handlers, log *zap.Logger) {
	app.Get("/api/health", h.Health.Check)

	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Ocealis API",
			"version": "1.0.0",
		})
	})

	v1 := app.Group("/api/v1")

	// User routes
	users := v1.Group("/users")
	users.Post("/", middleware.StrictRateLimit(), h.User.CreateUser)
	users.Get("/:id", middleware.Auth(), h.User.GetUser)

	// Bottle routes
	bottles := v1.Group("/bottles", middleware.Auth())
	bottles.Post("/", middleware.StrictRateLimit(), h.Bottle.CreateBottle)
	bottles.Get("/:id", h.Bottle.GetBottle)
	bottles.Get("/:id/journey", h.Bottle.GetJourney)
	bottles.Get("/:id/events", h.Event.GetBottleEvents)
	bottles.Get("/:id/discover", h.Bottle.DiscoverBottle)
	bottles.Get("/:id/release", h.Bottle.ReleaseBottle)
}
