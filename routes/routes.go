package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/services"
)

// SetupRouter initializes the router with all routes
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Initialize services
	userService := services.NewUserService()
	authService := services.NewAuthService()
	logService := services.NewLogService()
	monitoringService := services.NewMonitoringService()

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	authController := controllers.NewAuthController(authService)
	logController := controllers.NewLogController(logService)
	monitoringController := controllers.NewMonitoringController(monitoringService)

	// Define routes
	RegisterAuthRoutes(r, authController)
	RegisterAdminRoutes(r, userController, logController, monitoringController)
	RegisterUserRoutes(r, userController)

	return r
}
