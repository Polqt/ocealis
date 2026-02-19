package main

import (
	"time"

	"github.com/Polqt/ocealis/db"
	dbGen "github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/util"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Logger
	log, _ := zap.NewProduction()
	defer log.Sync()

	// Config
	_ = godotenv.Load()

	if err := db.Connect(); err != nil {
		log.Fatal("DB connection error", zap.Error(err))
	}

	defer db.Pool.Close()

	queries := dbGen.New(db.Pool)

	// Repositories

	// Services

	// Websocket

	// Scheduler

	// Handlers

	// Fiber app
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

	// Routes
	

	// Shutdown

}
