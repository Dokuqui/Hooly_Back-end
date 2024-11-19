package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/services"
)

func ParkingSpotRoutes(router *gin.Engine) {
	parkingSpotService := services.NewParkingSpotService()
	parkingSpotController := controllers.NewParkingSpotController(parkingSpotService)

	parking := router.Group("/parkingspots")
	{
		parking.GET("/", parkingSpotController.ListAllParkingSpots)                    // List all spots or filter by day
		parking.GET("/:id/available", parkingSpotController.IsSpotAvailable)           // Check availability
		parking.PUT("/:id/reservation", parkingSpotController.UpdateReservationStatus) // Update reservation status
	}
}
