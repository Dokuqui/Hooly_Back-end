package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/services"
	"gitlab.com/hooly2/back/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type ParkingSpotController struct {
	ParkingSpotServices *services.ParkingSpotService
}

// NewParkingSpotController creates a new instance of the controller
func NewParkingSpotController(parkingSpotController *services.ParkingSpotService) *ParkingSpotController {
	return &ParkingSpotController{ParkingSpotServices: parkingSpotController}
}

// ListAllParkingSpots handles GET requests to list all parking spots or filter by day
func (ctrl *ParkingSpotController) ListAllParkingSpots(c *gin.Context) {
	dayOfWeek := c.Query("day_of_week") // Optional query parameter

	spots, err := ctrl.ParkingSpotServices.ListAllParkingSpots(dayOfWeek, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parking spots"})
		return
	}

	c.JSON(http.StatusOK, spots)
}

func (ctrl *ParkingSpotController) CreateParkingSpotHandler(c *gin.Context) {
	// Ensure the user is an admin
	userRole, _ := c.Get("role")
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Get the day of the week from the request body
	var parkingSpot model.ParkingSpot
	if err := c.ShouldBindJSON(&parkingSpot); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Ensure the day_of_week is valid
	if !utils.IsValidDayOfWeek(parkingSpot.Day) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day of week"})
		return
	}

	// Call the service to create the parking spot
	createdSpot, err := ctrl.ParkingSpotServices.CreateParkingSpot(parkingSpot.Day, context.Background())
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Return the created parking spot as JSON
	c.JSON(http.StatusOK, gin.H{
		"message":      "Parking spot created successfully",
		"parking_spot": createdSpot,
	})
}

// IsSpotAvailable handles GET requests to check parking spot availability
func (ctrl *ParkingSpotController) IsSpotAvailable(c *gin.Context) {
	id := c.Param("id")
	spotID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid spot ID"})
		return
	}

	available, err := ctrl.ParkingSpotServices.IsSpotAvailable(spotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"available": available})
}

// UpdateReservationStatus handles PUT requests to update reservation status of a parking spot
func (ctrl *ParkingSpotController) UpdateReservationStatus(c *gin.Context) {
	id := c.Param("id")
	spotID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid spot ID"})
		return
	}

	var body struct {
		Reserved bool `json:"reserved"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = ctrl.ParkingSpotServices.UpdateReservationStatus(spotID, body.Reserved)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation status updated successfully"})
}
