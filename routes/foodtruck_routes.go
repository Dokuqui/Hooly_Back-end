package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
)

func RegisterFoodtruckRoutes(r *gin.Engine, foodtruckController *controllers.FoodtruckController) {

	foodtruck := r.Group("/foodtrucks")
	{
		foodtruck.GET("/:id/", foodtruckController.GetFoodtruckByID)
		foodtruck.GET("/", foodtruckController.GetFoodtrucksByName)
		foodtruck.POST("/add", foodtruckController.CreateFoodtruck)
		foodtruck.PUT("/:id", foodtruckController.UpdateFoodtruck)
		foodtruck.DELETE("/:id", foodtruckController.DeleteFoodtruck)
	}
}
