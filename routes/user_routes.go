package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/middleware"
)

// RegisterUserRoutes defines user-specific routes
func RegisterUserRoutes(api *gin.RouterGroup, userController *controllers.UserController) {
	user := api.Group("/user")
	user.Use(middleware.AuthMiddleware())
	{
		// User routes
		user.GET("/:id", userController.GetUserDetails)
		user.PUT("/:id", userController.UpdateUserDetails)
	}
}
