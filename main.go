package main

import (
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/routes"
	"log"
)

func main() {
	// Connect to MongoDB
	db.Connect()

	// Set up routes
	r := routes.SetupRouter()

	// Run the server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
