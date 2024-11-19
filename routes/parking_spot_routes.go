package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
)

func RegisterParkingSpotRoutes(r *gin.Engine, parkingSpotController *controllers.ParkingSpotController) {

	parking := r.Group("/parkingspots")
	{
		parking.GET("/", parkingSpotController.ListAllParkingSpots)                    // List all spots or filter by day
		parking.GET("/:id/available", parkingSpotController.IsSpotAvailable)           // Check availability
		parking.PUT("/:id/reservation", parkingSpotController.UpdateReservationStatus) // Update reservation status
	}
}
