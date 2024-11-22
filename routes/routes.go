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
	// Initialize the main gin.Engine router
	r := gin.Default()

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Get allowed origins for CORS from environment variables
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	// Configure CORS settings
	config := cors.Config{
		AllowOrigins:     []string{allowedOrigins},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}

	// Apply CORS middleware
	r.Use(cors.New(config))

	// Initialize services
	userService := services.NewUserService()
	authService := services.NewAuthService()
	logService := services.NewLogService()
	monitoringService := services.NewMonitoringService()
	foodtruckService := services.NewFoodtruckService()
	parkingSpotService := services.NewParkingSpotService()
	reservationService := services.NewReservationService()

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	authController := controllers.NewAuthController(authService)
	logController := controllers.NewLogController(logService)
	monitoringController := controllers.NewMonitoringController(monitoringService)
	foodtruckController := controllers.NewFoodtruckController(foodtruckService)
	parkingSpotController := controllers.NewParkingSpotController(parkingSpotService)
	reservationController := controllers.NewReservationController(reservationService)

	// Define a route group for '/api'
	api := r.Group("/api") // Create a group for '/api'
	{
		// Register all the routes under the '/api' group
		RegisterAuthRoutes(api, authController)                                       // Use *gin.Engine
		RegisterAdminRoutes(api, userController, logController, monitoringController) // Use *gin.Engine
		RegisterUserRoutes(api, userController)                                       // Use *gin.Engine
		RegisterFoodtruckRoutes(api, foodtruckController)                             // Use *gin.Engine
		RegisterParkingSpotRoutes(api, parkingSpotController)                         // Use *gin.Engine
		RegisterReservationRoutes(api, reservationController)                         // Use *gin.Engine
	}

	// Return the main Gin router object, which is *gin.Engine
	return r

}
