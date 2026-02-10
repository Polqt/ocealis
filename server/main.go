package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Polqt/ocealis/api"
	"github.com/Polqt/ocealis/db"
	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/services"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	if err := db.Connect(); err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	defer db.Pool.Close()

	fmt.Println("Server starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queries := ocealis.New(db.Pool)

	// Initialize services
	bottleService := services.NewBottleService(queries)
	// driftService := services.NewDriftService(queries)

	// Initialize handlers
	healthHandler := api.NewHealthHandler(queries)
	bottleHandler := api.NewBottleHandler(bottleService)

	// Initialize fiber
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Ocealis",
		AppName:       "Ocealis v1",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{os.Getenv("CORS_ALLOWED_ORIGINS")},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	// Routes
	apiV1 := app.Group("/api")

	api.RegisterHealthRoutes(apiV1, healthHandler)
	api.RegisterBottleRoutes(apiV1, bottleHandler)

	// Get port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("Shutting down...")
		cancel()
		_ = app.Shutdown()
	}()

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server:%v", err)
	}

	// Root Endpoint
	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Ocealis API",
			"version": "1.0.0",
			"status": "running",
			"endpoints": fiber.Map{
				"health": "GET /api/health",
				"bottles": "GET /api/bottles",
			},
		})
	})
}

