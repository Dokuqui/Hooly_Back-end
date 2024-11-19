package controllers

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/services"
	"log"
	"net/http"
)

type AuthController struct {
	AuthService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{AuthService: authService}
}

// Signup handles user registration
func (ac *AuthController) Signup(c *gin.Context) {
	var user model.User

	// Bind the incoming JSON body to the User model
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	// Call the AuthService Signup function
	newUser, err := ac.AuthService.Signup(user.Email, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with the newly created user (excluding password)
	c.JSON(http.StatusOK, gin.H{"user": newUser})
}

// Login handles user login
func (ac *AuthController) Login(c *gin.Context) {
	var user model.User

	// Bind the incoming JSON body to the User model
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	// Call the AuthService Login function
	token, err := ac.AuthService.Login(user.Email, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with the JWT token
	c.JSON(http.StatusOK, gin.H{"token": token})
}
