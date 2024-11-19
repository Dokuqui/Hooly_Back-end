package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// RoleMiddleware checks the role of the user from the JWT token and ensures they have the required role for the route
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the current user's role from the JWT token (should be set by your authentication middleware)
		currentRole := c.GetString("role")

		// Check if the current user's role matches the required role
		if currentRole != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			c.Abort() // Prevent further processing
			return
		}

		// If the role matches, continue with the next handler
		c.Next()
	}
}
