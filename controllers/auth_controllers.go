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

	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	newUser, token, err := ac.AuthService.Signup(user.Email, user.Firstname, user.Lastname, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with the user details and JWT token
	c.JSON(http.StatusOK, gin.H{
		"user":  gin.H{"id": newUser.ID, "email": newUser.Email, "firstname": newUser.Firstname, "lastname": newUser.Lastname},
		"token": token,
	})
}

// Login handles user login
func (ac *AuthController) Login(c *gin.Context) {
	var user model.User

	// Bind the incoming JSON body to the loginRequest struct
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	// Call the AuthService Login function
	token, err := ac.AuthService.Login(user.Email, user.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Respond with the JWT token and user details
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID.Hex(),
			"email":     user.Email,
			"firstname": user.Firstname,
			"lastname":  user.Lastname,
			"role":      user.Role,
		},
	})
}
