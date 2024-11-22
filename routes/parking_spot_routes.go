package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/middleware"
)

func RegisterParkingSpotRoutes(api *gin.RouterGroup, parkingSpotController *controllers.ParkingSpotController) {

	parking := api.Group("/parkingspots", middleware.AuthMiddleware())
	{
		parking.GET("/", parkingSpotController.ListAllParkingSpots)
		parking.PUT("/:id/reservation", parkingSpotController.UpdateReservationStatus)
		parking.POST("/create", parkingSpotController.CreateParkingSpotHandler)
	}
}
