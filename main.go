package main

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/db"
	"log"
)

func main() {
	// Connect to MongoDB
	db.Connect()

	// Initialize Gin router
	r := gin.Default()

	// Test route to confirm MongoDB connection
	r.GET("/test-db", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "MongoDB connection successful!",
		})
	})

	// Run the server on port 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
