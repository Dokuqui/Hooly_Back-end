package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/services"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type UserController struct {
	UserServices *services.UserService
}

func NewUserController(userController *services.UserService) *UserController {
	return &UserController{UserServices: userController}
}

// CreateUser handles creating a new user
func (uc *UserController) CreateUser(c *gin.Context) {
	// Extract the current user's role from the JWT token
	currentRole := c.GetString("role")
	if currentRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// Bind the request body to a struct
	var userInput struct {
		FirstName string     `json:"first_name" binding:"required"`
		LastName  string     `json:"last_name" binding:"required"`
		Email     string     `json:"email" binding:"required,email"`
		Password  string     `json:"password" binding:"required,min=6"`
		Role      model.Role `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the UserService to create the new user
	newUser, err := uc.UserServices.CreateUser(
		userInput.FirstName,
		userInput.LastName,
		userInput.Email,
		userInput.Password,
		userInput.Role,
	)
	if err != nil {
		if err.Error() == "email already registered" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Respond with the created user (excluding the password)
	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":         newUser.ID.Hex(),
			"first_name": newUser.Firstname,
			"last_name":  newUser.Lastname,
			"email":      newUser.Email,
			"role":       newUser.Role,
		},
	})
}

// GetAllUsers fetches the list of all users
func (uc *UserController) GetAllUsers(c *gin.Context) {
	// Extract the current user's role from the JWT token (JWT middleware will set this in context)
	currentRole := c.GetString("role")

	// Ensure that only admin users can access this endpoint
	if currentRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// Call UserService to fetch all users from DB
	users, err := uc.UserServices.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the list of users
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// GetUserDetails fetches details of a specific user
func (uc *UserController) GetUserDetails(c *gin.Context) {
	// Get the user ID from the URL parameter
	userID := c.Param("id")

	// Call UserService to fetch the user details from DB
	user, err := uc.UserServices.GetUserById(userID)
	if errors.Is(mongo.ErrNoDocuments, err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the user details (without password)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUserDetails updates the details of a specific user
func (uc *UserController) UpdateUserDetails(c *gin.Context) {
	// Extract user ID from URL parameter
	userID := c.Param("id")

	// Extract the current user's ID and role from the JWT token (role set by middleware)
	currentUserID := c.GetString("user_id")
	currentRole := c.GetString("role")

	// Ensure that the current user is either the user themselves or an admin
	if currentUserID != userID && currentRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// Bind the incoming request body to the updated user data
	var updatedUser model.User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call UserService to update the user in the database
	updatedUserResult, err := uc.UserServices.UpdateUser(userID, updatedUser)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Respond with the updated user data
	c.JSON(http.StatusOK, gin.H{"user": updatedUserResult})
}

// DeleteUser deletes a specific user
func (uc *UserController) DeleteUser(c *gin.Context) {
	// Get the user ID from the URL parameter
	userID := c.Param("id")

	// Get the current user's ID and role from the JWT token (JWT middleware will add these)
	currentUserID := c.GetString("user_id")
	currentRole := c.GetString("role")

	// Check if the current user is an admin or trying to delete their own account
	if currentUserID != userID && currentRole != "admin" { // Allow deletion only if it's the same user or admin
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// Call UserService to delete the user from the database
	err := uc.UserServices.DeleteUser(userID)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
