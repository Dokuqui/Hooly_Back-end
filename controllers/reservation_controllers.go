package controllers

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/services"
	"gitlab.com/hooly2/back/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
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
	// Extract the current user's role from the JWT token (JWT middleware will set this in context)
	currentRole := ctx.GetString("role")

	// Ensure that only admin users can access this endpoint
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

// GetUserReservationsHandler retrieves reservations for the logged-in user.
func (c *ReservationController) GetUserReservationsHandler(ctx *gin.Context) {
	// Get user_id from context (set during authentication)
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

	// Get user_id from context (set during authentication)
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

	// Bind the JSON body to the reservation struct
	if err := ctx.ShouldBindJSON(&reservation); err != nil {
		log.Println("Failed to bind request body:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Convert SpotID to ObjectID
	spotObjectID, err := primitive.ObjectIDFromHex(reservation.SpotID.Hex())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SpotID"})
		return
	}
	reservation.SpotID = spotObjectID

	// Retrieve the userID from the context (set by JWT middleware)
	userID, exists := ctx.Get("userId")
	if !exists {
		log.Println("User ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID missing"})
		return
	}

	// Convert the userID to a primitive.ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		log.Println("Invalid user ID:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	reservation.UserID = objectID

	// Validate that SpotID is a valid ObjectID
	if !reservation.SpotID.IsZero() {
		if _, err := primitive.ObjectIDFromHex(reservation.SpotID.Hex()); err != nil {
			log.Println("Invalid Spot ID:", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Spot ID"})
			return
		}
	}

	// Call the service to create the reservation
	err = c.ReservationService.CreateReservation(ctx, &reservation)
	if err != nil {
		// Log error and respond based on the error message
		log.Println("Failed to create reservation:", err)

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
		log.Println("Reservation ID is empty, something went wrong")
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
	// Check if user is admin
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
func (c *ReservationController) GetAllReservedSlotsHandler(ctx *gin.Context) {
    slots, err := c.ReservationService.GetAllFutureReservations(ctx.Request.Context())
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "error": "Erreur lors de la récupération des créneaux réservés",
            "details": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "data": slots,
    })
}
