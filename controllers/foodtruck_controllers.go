package controllers

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/services"
	"gitlab.com/hooly2/back/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type FoodtruckController struct {
	FoodtruckServices *services.FoodtruckService
}

func NewFoodtruckController(foodtruckController *services.FoodtruckService) *FoodtruckController {
	return &FoodtruckController{FoodtruckServices: foodtruckController}
}

// CreateFoodtruck Add a foodtruck
func (c *FoodtruckController) CreateFoodtruck(ctx *gin.Context) {
	var foodtruck model.Foodtruck

	// Bind the incoming JSON to the Foodtruck struct
	if err := ctx.ShouldBindJSON(&foodtruck); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Retrieve the userID from the context, set by the JWT middleware
	userID, exists := ctx.Get("userId") // Ensure the key matches your middleware
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert the userID to a primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	foodtruck.UserID = objectID

	// Call the service to add the food truck
	addedFoodtruck, err := c.FoodtruckServices.AddFoodtruck(&foodtruck)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create food truck"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":        addedFoodtruck.ID.Hex(),
		"userID":    addedFoodtruck.UserID.Hex(),
		"foodtruck": addedFoodtruck.Name,
	})
}

// GetAllFoodTrucks retrieves all food trucks (admin only).
func (c *FoodtruckController) GetAllFoodTrucks(ctx *gin.Context) {
	// Extract the current user's role from the JWT token
	currentRole := ctx.GetString("role")

	// Ensure only admin users can access this endpoint
	if currentRole != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	foodtrucks, err := c.FoodtruckServices.GetAllFoodTrucks(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": foodtrucks})
}

// GetUserFoodTrucksHandler retrieves all food trucks associated with the authenticated user
func (c *FoodtruckController) GetUserFoodTrucksHandler(ctx *gin.Context) {
	// Get user_id from context (set during authentication)
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	foodtrucks, err := c.FoodtruckServices.GetUserFoodTrucks(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user foodtrucks"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": foodtrucks})
}

// GetFoodtruckByIDHandler retrieves a foodtruck by ID (scoped by user ID).
func (c *FoodtruckController) GetFoodtruckByIDHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	foodtruckID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	foodtruck, err := c.FoodtruckServices.GetFoodTruckByID(ctx, foodtruckID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": foodtruck})
}

// UpdateFoodtruck Update a foodtruck
func (c *FoodtruckController) UpdateFoodtruck(ctx *gin.Context) {
	id := ctx.Param("id")
	foodtruckID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var updateData bson.M
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	if err := c.FoodtruckServices.UpdateFoodtruck(ctx, foodtruckID, updateData, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Food truck updated successfully"})
}

// DeleteFoodtruck deletes a foodtruck by ID.
func (c *FoodtruckController) DeleteFoodtruck(ctx *gin.Context) {
	id := ctx.Param("id")
	foodtruckID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	if err := c.FoodtruckServices.DeleteFoodtruck(ctx, foodtruckID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Foodtruck deleted"})
}

// AdminDeleteFoodtruck deletes any foodtruck (admin only).
func (c *FoodtruckController) AdminDeleteFoodtruck(ctx *gin.Context) {
	userRole, _ := ctx.Get("role")
	if userRole != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Parse reservation ID from URL
	id := ctx.Param("id")
	foodtruckID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation ID"})
		return
	}

	err = c.FoodtruckServices.AdminDeleteFoodtruck(ctx, foodtruckID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete foodtruck"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Foodtruck deleted successfully"})
}
