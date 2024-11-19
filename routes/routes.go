package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/services"
	"log"
	"os"
)

// SetupRouter initializes the router with all routes
func SetupRouter() *gin.Engine {
	r := gin.Default()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	config := cors.Config{
		AllowOrigins:     []string{allowedOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}

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

	r.Use(cors.New(config))

	return r
}
