package main

import (
	"context"
	"errors"
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
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewDevelopment()
	defer func() {
		_ = log.Sync()
	}()

	if err := db.Connect(log); err != nil {
		log.Fatal("database connection error", zap.Error(err))
	}
	defer db.Pool.Close()

	queries := dbGen.New(db.Pool)

	bottleRepo := repository.NewBottleRepository(queries)
	eventRepo := repository.NewEventRepository(queries)
	userRepo := repository.NewUserRepository(queries)

	hub := ws.NewHub()
	broadcaster := ws.NewBroadcaster(hub, log)

	bottleSvc := service.NewBottleService(db.Pool, bottleRepo, eventRepo, broadcaster)
	userSvc := service.NewUserService(userRepo)
	driftSvc := service.NewDriftService(db.Pool, bottleRepo, eventRepo, broadcaster, log)
	discoverySvc := service.NewDiscoveryService(bottleRepo)

	h := api.Handlers{
		Health:    handler.NewHealthHandler(db.Pool, hub),
		Bottle:    handler.NewBottleHandler(bottleSvc),
		User:      handler.NewUserHandler(userSvc),
		Event:     handler.NewEventHandler(eventRepo),
		Discovery: handler.NewDiscoveryHandler(discoverySvc),
	}

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

	api.RegisterRoutes(app, h, hub, log)

	scheduler := service.NewScheduler(driftSvc, log)
	appCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	scheduler.Start(appCtx)
	defer scheduler.Stop()

	port := util.EnvString("PORT", "8080")

	go func() {
		log.Info("server starting", zap.String("port", port))
		if err := app.Listen(":" + port); err != nil && !errors.Is(err, fiber.ErrServiceUnavailable) {
			log.Error("server listen error", zap.Error(err))
		}
	}()

	<-appCtx.Done()
	log.Info("shutdown signal received, draining connections")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("server shutdown error", zap.Error(err))
	}

	log.Info("server exited cleanly")
}
