package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	// TODO: rework server init

	r := gin.Default()

	// Define a simple GET route for testing
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Hooly Parking Reservation System",
		})
	})

	// Run the server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
