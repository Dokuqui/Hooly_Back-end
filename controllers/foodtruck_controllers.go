package controllers

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type FoodtruckController struct {
	FoodtruckServices *services.FoodtruckService
}

func NewFoodtruckController(foodtruckController *services.FoodtruckService) *FoodtruckController {
	return &FoodtruckController{FoodtruckServices: foodtruckController}
}

// Add a foodtruck
func (c *FoodtruckController) CreateFoodtruck(ctx *gin.Context) {
	var foodtruck model.Foodtruck
	if err := ctx.ShouldBindJSON(&foodtruck); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	foodtruck.UserID = userID.(primitive.ObjectID)

	added, err := c.FoodtruckServices.AddFoodtruck(&foodtruck)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create food truck"})
		return
	}

	ctx.JSON(http.StatusCreated, added)
}

// Get foodtruck by ID
func (c *FoodtruckController) GetFoodtruckByID(ctx *gin.Context) {
	id := ctx.Param("id")

	// Extract userID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	foodtruck, err := c.FoodtruckServices.FindFoodtruckByID(userID.(string), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Food truck not found"})
		return
	}

	ctx.JSON(http.StatusOK, foodtruck)
}

// Get foodtruck by NAME
func (c *FoodtruckController) GetFoodtrucksByName(ctx *gin.Context) {
	name := ctx.Query("name")

	// Extract userID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	foodtrucks, err := c.FoodtruckServices.FindFoodtruckByName(userID.(string), name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch food trucks"})
		return
	}

	ctx.JSON(http.StatusOK, foodtrucks)
}

// GetAllFoodtrucksByUserID retrieves all food trucks associated with the authenticated user
func (c *FoodtruckController) GetAllFoodtrucksByUserID(ctx *gin.Context) {
	// Extract userID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Fetch all food trucks for the user
	foodtrucks, err := c.FoodtruckServices.GetAllFoodtrucksByUserID(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch food trucks"})
		return
	}

	ctx.JSON(http.StatusOK, foodtrucks)
}

// Update a foodtruck
func (c *FoodtruckController) UpdateFoodtruck(ctx *gin.Context) {
	id := ctx.Param("id")

	var foodtruck model.Foodtruck

	if err := ctx.ShouldBindJSON(&foodtruck); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.FoodtruckServices.UpdateFoodtruck(id, userID.(primitive.ObjectID), &foodtruck); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update food truck"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Food truck updated successfully"})
}

// Delete a foodtruck
func (c *FoodtruckController) DeleteFoodtruck(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.FoodtruckServices.DeleteFoodtruck(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete food truck"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Food truck deleted successfully"})
}
