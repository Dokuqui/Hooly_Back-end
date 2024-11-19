package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
)

// RegisterAuthRoutes defines authentication-related routes
func RegisterAuthRoutes(r *gin.Engine, authController *controllers.AuthController) {
	r.POST("/signup", authController.Signup)
	r.POST("/login", authController.Login)
}
