package main

import (
	"fmt"
	"log"

	"github.com/Polqt/ocealis/db"
)

func main() {
	if err := db.Connect(); err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	defer db.Close()

	fmt.Println("Server starting...")

}
