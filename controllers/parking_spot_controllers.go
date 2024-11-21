package controllers

import (
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

	spots, err := ctrl.ParkingSpotServices.ListAllParkingSpots(dayOfWeek, c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parking spots", "details": err.Error()})
		return
	}

	if len(spots) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No parking spots found"})
		return
	}

	c.JSON(http.StatusOK, spots)
}

// CreateParkingSpotHandler handles POST requests to create a new parking spot
func (ctrl *ParkingSpotController) CreateParkingSpotHandler(c *gin.Context) {
	// Ensure the user is an admin
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Parse input data
	var parkingSpot model.ParkingSpot
	if err := c.ShouldBindJSON(&parkingSpot); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Validate day_of_week
	if !utils.IsValidDayOfWeek(parkingSpot.Day) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid day of week"})
		return
	}

	// Call the service to create the parking spot
	createdSpot, err := ctrl.ParkingSpotServices.CreateParkingSpot(parkingSpot.Day, c.Request.Context())
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Parking spot created successfully",
		"parking_spot": createdSpot,
	})
}

// UpdateReservationStatus handles PUT requests to update reservation status of a parking spot
func (ctrl *ParkingSpotController) UpdateReservationStatus(c *gin.Context) {
	id := c.Param("id")

	// Validate the ObjectID
	spotID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid spot ID"})
		return
	}

	// Parse request body
	var body struct {
		Reserved bool `json:"reserved"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Update reservation status via the service
	err = ctrl.ParkingSpotServices.UpdateReservationStatus(spotID, body.Reserved, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reservation status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation status updated successfully"})
}
