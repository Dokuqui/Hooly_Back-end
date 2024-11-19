package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/middleware"
	"gitlab.com/hooly2/back/services"
)

// UserRoutes sets up all the user-related routes
func UserRoutes(r *gin.Engine) {
	// Initialize the user service and controller
	userService := services.NewUserService()
	authService := services.NewAuthService()
	userController := controllers.NewUserController(userService)
	authController := controllers.NewAuthController(authService)

	// Define user-related routes
	r.POST("/signup", authController.Signup)
	r.POST("/login", authController.Login)

	// Group routes that need JWT and role-based access
	authorized := r.Group("/")
	authorized.Use(middleware.JWTMiddleware())

	// Admin-only routes
	admin := authorized.Group("/")
	admin.Use(middleware.RoleMiddleware("admin"))
	{
		admin.GET("/users", userController.GetAllUsers)
		admin.DELETE("/users/:id", userController.DeleteUser)
	}

	// User-specific routes
	user := authorized.Group("/")
	{
		user.GET("/users/:id", userController.GetUserDetails)
		user.PUT("/users/:id", userController.UpdateUserDetails)
	}
}
