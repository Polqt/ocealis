package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Polqt/ocealis/db"
	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	if err := db.Connect(); err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	defer db.Pool.Close()

	fmt.Println("Server starting...")

	queries := ocealis.New(db.Pool)
	users, err := queries.GetUser(context.Background(), 1)
	if err != nil {
		log.Fatalf("Failed to query users: %v", err)
	}
	fmt.Printf("Users in database: %v\n", users)

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
}
