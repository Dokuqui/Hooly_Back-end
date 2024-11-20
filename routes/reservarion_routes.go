package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
)

func RegisterReservationRoutes(r *gin.Engine, reservationController *controllers.ReservationController) {

	reservation := r.Group("/reservation")
	{
		reservation.POST("", reservationController.CreateReservationHandler)
		reservation.PUT("/:id", reservationController.UpdateReservationHandler)
		reservation.DELETE("/:id", reservationController.DeleteReservationHandler)
	}
}
