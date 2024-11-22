package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/middleware"
)

func RegisterFoodtruckRoutes(r *gin.Engine, foodtruckController *controllers.FoodtruckController) {

	foodtruck := r.Group("/foodtrucks", middleware.AuthMiddleware())
	{
		foodtruck.GET("/", foodtruckController.GetAllFoodTrucks)
		foodtruck.GET("/:id", foodtruckController.GetFoodtruckByIDHandler)
		foodtruck.GET("/user", foodtruckController.GetUserFoodTrucksHandler)
		foodtruck.POST("/add", foodtruckController.CreateFoodtruck)
		foodtruck.PUT("/:id", foodtruckController.UpdateFoodtruck)
		foodtruck.DELETE("/:id", foodtruckController.DeleteFoodtruck)
		foodtruck.DELETE("/admin/:id", foodtruckController.AdminDeleteFoodtruck)
	}
}
