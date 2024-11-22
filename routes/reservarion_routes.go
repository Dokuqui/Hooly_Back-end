package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/middleware"
)

func RegisterReservationRoutes(api *gin.RouterGroup, reservationController *controllers.ReservationController) {

	reservation := api.Group("/reservation", middleware.AuthMiddleware())
	{
		reservation.GET("/admin", reservationController.GetAllReservationsHandler)
		reservation.GET("/:id", reservationController.GetReservationByIDHandler)
		reservation.DELETE("/admin/:id", reservationController.AdminDeleteReservationHandler)
		reservation.POST("/", reservationController.CreateReservationHandler)
		reservation.PUT("/:id", reservationController.UpdateReservationHandler)
		reservation.DELETE("/:id", reservationController.DeleteReservationHandler)
		reservation.GET("/user", reservationController.GetUserReservationsHandler)
		reservation.GET("/users", reservationController.GetAllUserReservationsHandler)
		reservation.GET("/user/:id", reservationController.GetReservationByIDHandler)
	}
}
