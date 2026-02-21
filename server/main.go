package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Polqt/ocealis/api"
	"github.com/Polqt/ocealis/api/handler"
	"github.com/Polqt/ocealis/db"
	dbGen "github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/internal/service"
	"github.com/Polqt/ocealis/util"
	"github.com/Polqt/ocealis/ws"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Logger — structured JSON in production.
	log, _ := zap.NewProduction()
	defer log.Sync()

	// Config — .env is optional in deployed envs (already set via the platform).
	_ = godotenv.Load()

	// Database pool.
	if err := db.Connect(); err != nil {
		log.Fatal("DB connection error", zap.Error(err))
	}
	defer db.Pool.Close()

	queries := dbGen.New(db.Pool)

	// Repositories.
	bottleRepo := repository.NewBottleRepository(queries)
	eventRepo := repository.NewEventRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	// WebSocket hub + broadcaster (created before services so they can broadcast).
	hub := ws.NewHub()
	broadcaster := ws.NewBroadcaster(hub, log)

	// Services.
	bottleSvc := service.NewBottleService(bottleRepo, eventRepo, broadcaster)
	userSvc := service.NewUserService(userRepo)
	driftSvc := service.NewDriftService(bottleRepo, eventRepo, broadcaster, log)

	// Background scheduler — drift tick every 15 minutes.
	scheduler := service.NewScheduler(driftSvc, log)
	scheduler.Start(context.Background())
	defer scheduler.Stop()

	// HTTP handlers.
	h := api.Handlers{
		Health: handler.NewHealthHandler(),
		Bottle: handler.NewBottleHandler(bottleSvc),
		User:   handler.NewUserHandler(userSvc),
		Event:  handler.NewEventHandler(eventRepo),
	}

	// Fiber app.
	app := fiber.New(fiber.Config{
		AppName:       "Ocealis v1",
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Ocealis",
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "Internal Server Error"
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
				msg = e.Message
			}
			return c.Status(code).JSON(fiber.Map{"error": msg})
		},
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{util.EnvString("CORS_ALLOWED_ORIGINS", "http://localhost:3000")},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// Routes (HTTP + WebSocket).
	api.RegisterRoutes(app, h, hub, log)

	// Shutdown: wait for SIGINT/SIGTERM before draining connections.
	port := util.EnvString("PORT", "8080")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("server starting", zap.String("port", port))
		if err := app.Listen(":" + port); err != nil {
			log.Error("server listen error", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutdown signal received, draining connections")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error("server shutdown error", zap.Error(err))
	}
	log.Info("server exited cleanly")
}
