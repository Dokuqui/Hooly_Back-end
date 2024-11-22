package controllers

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/services"
	"gitlab.com/hooly2/back/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strings"
)

type ReservationController struct {
	ReservationService *services.ReservationService
}

func NewReservationController(reservationService *services.ReservationService) *ReservationController {
	return &ReservationController{ReservationService: reservationService}
}

// GetAllReservationsHandler retrieves all reservations (Admin only).
func (c *ReservationController) GetAllReservationsHandler(ctx *gin.Context) {
	currentRole := ctx.GetString("role")

	if currentRole != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	reservations, err := c.ReservationService.GetAllReservations(ctx)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": reservations})
}

// GetAllUserReservationsHandler retrieves all reservations for the current user (without userID and foodTruckID).
func (c *ReservationController) GetAllUserReservationsHandler(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	// Fetch all reservations for the user
	reservations, err := c.ReservationService.GetUserReservations(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the reservations without sensitive information
	ctx.JSON(http.StatusOK, gin.H{"data": reservations})
}

// GetUserReservationsHandler retrieves reservations for the logged-in user.
func (c *ReservationController) GetUserReservationsHandler(ctx *gin.Context) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	reservations, err := c.ReservationService.GetUserReservations(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user reservations"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": reservations})
}

// GetReservationByIDHandler retrieves a reservation by ID (scoped by user ID).
func (c *ReservationController) GetReservationByIDHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	reservationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	reservation, err := c.ReservationService.GetReservationByID(ctx, reservationID, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": reservation})
}

// CreateReservationHandler creates a new reservation.
func (c *ReservationController) CreateReservationHandler(ctx *gin.Context) {
	var reservation model.Reservation

	if err := ctx.ShouldBindJSON(&reservation); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	spotObjectID, err := primitive.ObjectIDFromHex(reservation.SpotID.Hex())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SpotID"})
		return
	}
	reservation.SpotID = spotObjectID

	// Retrieve the userID from the context (set by JWT middleware)
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID missing"})
		return
	}

	// Convert the userID to a primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	reservation.UserID = objectID

	// Validate that SpotID is a valid ObjectID
	if !reservation.SpotID.IsZero() {
		if _, err := primitive.ObjectIDFromHex(reservation.SpotID.Hex()); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Spot ID"})
			return
		}
	}

	// Call the service to create the reservation
	err = c.ReservationService.CreateReservation(ctx, &reservation)
	if err != nil {
		// Check for specific error messages to send a 400 Bad Request
		if strings.Contains(err.Error(), "spot is not available") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Spot is not available"})
		} else if strings.Contains(err.Error(), "already has a reservation") ||
			strings.Contains(err.Error(), "no available spots") ||
			strings.Contains(err.Error(), "past date or today") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		}
		return
	}

	// Ensure the reservation ID is set correctly after insertion
	if reservation.ID.IsZero() {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	// Respond with the created reservation details
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Reservation created successfully",
		"reservation": gin.H{
			"id":            reservation.ID.Hex(),
			"spot_id":       reservation.SpotID.Hex(),
			"food_truck_id": reservation.FoodTruckID.Hex(),
			"user_id":       reservation.UserID.Hex(),
			"date":          reservation.Date,
			"created_at":    reservation.CreatedAt,
		},
	})
}

// UpdateReservationHandler updates an existing reservation.
func (c *ReservationController) UpdateReservationHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	reservationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var updateData bson.M
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user_id from context (set during authentication)
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	if err := c.ReservationService.UpdateReservation(ctx, reservationID, updateData, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "reservation updated"})
}

// DeleteReservationHandler deletes a reservation by ID.
func (c *ReservationController) DeleteReservationHandler(ctx *gin.Context) {
	id := ctx.Param("id")
	reservationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	if err := c.ReservationService.DeleteReservation(ctx, reservationID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "reservation deleted"})
}

// AdminDeleteReservationHandler deletes any reservation (admin only).
func (c *ReservationController) AdminDeleteReservationHandler(ctx *gin.Context) {
	userRole, _ := ctx.Get("role")
	if userRole != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Parse reservation ID from URL
	id := ctx.Param("id")
	reservationID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation ID"})
		return
	}

	// Call AdminDeleteReservation to handle the deletion and reserved count update
	err = c.ReservationService.AdminDeleteReservation(ctx, reservationID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond to the client with a success message
	ctx.JSON(http.StatusOK, gin.H{"message": "reservation deleted successfully"})
}
