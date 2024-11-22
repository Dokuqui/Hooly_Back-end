package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
)

// RegisterAuthRoutes defines authentication-related routes
func RegisterAuthRoutes(api *gin.RouterGroup, authController *controllers.AuthController) {
	api.POST("/signup", authController.Signup)
	api.POST("/login", authController.Login)
}
