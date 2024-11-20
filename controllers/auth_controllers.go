package controllers

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/services"
	"go.mongodb.org/mongo-driver/bson"
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

	// Bind incoming JSON to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	// Call the AuthService Signup method
	newUser, token, err := ac.AuthService.Signup(user.Email, user.Firstname, user.Lastname, user.Password)
	if err != nil {
		log.Println("Error during signup:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respond with user details and JWT token
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        newUser.ID.Hex(),
			"email":     newUser.Email,
			"firstname": newUser.Firstname,
			"lastname":  newUser.Lastname,
			"role":      newUser.Role,
		},
		"token": token,
	})
}

// Login handles user login
func (ac *AuthController) Login(c *gin.Context) {
	var loginRequest model.User

	// Bind the incoming JSON body to the loginRequest struct
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	// Call the AuthService Login function
	token, err := ac.AuthService.Login(loginRequest.Email, loginRequest.Password)
	if err != nil {
		log.Println("Error during login:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Get the user details from the database after successful login
	var user model.User
	err = ac.AuthService.UserCollection.FindOne(c, bson.M{"email": loginRequest.Email}).Decode(&user)
	if err != nil {
		log.Println("Error fetching user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user data"})
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
