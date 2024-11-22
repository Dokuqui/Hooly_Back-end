package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

// GetUserIDFromContext is a utility function to extract and validate user ID from the Gin context.
func GetUserIDFromContext(ctx *gin.Context) (primitive.ObjectID, error) {
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return primitive.ObjectID{}, fmt.Errorf("user_id not found in context")
	}

	// Convert user_id to string and then to primitive.ObjectID
	userIDStr, ok := userID.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user_id is not a valid string"})
		return primitive.ObjectID{}, fmt.Errorf("user_id is not a valid string")
	}

	userIDPrimitive, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user_id is not a valid ObjectID"})
		return primitive.ObjectID{}, fmt.Errorf("user_id is not a valid ObjectID")
	}

	return userIDPrimitive, nil
}
