package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/middleware"
)

func RegisterParkingSpotRoutes(r *gin.Engine, parkingSpotController *controllers.ParkingSpotController) {

	parking := r.Group("/parkingspots", middleware.AuthMiddleware())
	{
		parking.GET("/", parkingSpotController.ListAllParkingSpots)
		parking.GET("/:id/available", parkingSpotController.IsSpotAvailable)
		parking.PUT("/:id/reservation", parkingSpotController.UpdateReservationStatus)
		parking.POST("/create", parkingSpotController.CreateParkingSpotHandler)
	}
}
