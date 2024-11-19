package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/middleware"
)

// RegisterUserRoutes defines user-specific routes
func RegisterUserRoutes(r *gin.Engine, userController *controllers.UserController) {
	user := r.Group("/")
	user.Use(middleware.JWTMiddleware())
	{
		user.GET("/users/:id", userController.GetUserDetails)
		user.PUT("/users/:id", userController.UpdateUserDetails)
	}
}
