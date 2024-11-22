package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/middleware"
)

// RegisterAdminRoutes defines admin-only routes
func RegisterAdminRoutes(api *gin.RouterGroup, userController *controllers.UserController, logController *controllers.LogController, monitoringController *controllers.MonitoringController) {
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware("admin"))
	{
		admin.GET("/users", userController.GetAllUsers)
		admin.POST("/users", userController.CreateUser)
		admin.DELETE("/users/:id", userController.DeleteUser)

		// Log routes
		admin.POST("/logs", logController.CreateLogHandler)
		admin.GET("/logs", logController.FetchLogsHandler)

		// Monitoring routes
		admin.GET("/monitoring", monitoringController.FetchMonitoringDataHandler)
	}
}
