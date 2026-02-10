package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Polqt/ocealis/db"
	"github.com/Polqt/ocealis/db/ocealis"
)

func main() {
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

}
