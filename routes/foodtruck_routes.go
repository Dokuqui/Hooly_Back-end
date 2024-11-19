package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/controllers"
	"gitlab.com/hooly2/back/services"
)

func FoodtruckRoutes(router *gin.Engine) {
	foodtruckService := services.NewFoodtruckService()
	foodtruckController := controllers.NewFoodtruckController(foodtruckService)

	foodtruck := router.Group("/foodtrucks")
	{
		foodtruck.GET("/:id/", foodtruckController.GetFoodtruckByID)
		foodtruck.GET("/", foodtruckController.GetFoodtrucksByName)
		foodtruck.POST("/add", foodtruckController.CreateFoodtruck)
		foodtruck.PUT("/:id", foodtruckController.UpdateFoodtruck)
		foodtruck.DELETE("/:id", foodtruckController.DeleteFoodtruck)
	}
}
