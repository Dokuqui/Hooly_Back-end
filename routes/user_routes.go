package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/middleware"
)

// RegisterUserRoutes defines user-specific routes
func RegisterUserRoutes(r *gin.Engine, userController *controllers.UserController, reservationController *controllers.ReservationController) {
	user := r.Group("/")
	user.Use(middleware.JWTMiddleware())
	{
		// User routes
		user.GET("/users/:id", userController.GetUserDetails)
		user.PUT("/users/:id", userController.UpdateUserDetails)

		// Reservation routes
		user.GET("/reservation/user", reservationController.GetUserReservationsHandler)
		user.GET("/reservation/:id", reservationController.GetReservationByIDHandler)
	}
}
