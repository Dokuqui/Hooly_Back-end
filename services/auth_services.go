package services

import (
	"context"
	"errors"
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type AuthService struct {
	UserCollection *mongo.Collection
}

func NewAuthService() *AuthService {
	return &AuthService{
		UserCollection: db.GetCollection("user"),
	}
}

// Signup handles new user creation
func (s *AuthService) Signup(email, firstname, lastname, password string) (*model.User, string, error) {
	var existingUser model.User
	err := s.UserCollection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&existingUser)
	if err == nil {
		return nil, "", errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", errors.New("failed to hash password")
	}

	newUser := model.User{
		ID:        primitive.NewObjectID(),
		Email:     email,
		Firstname: firstname,
		Lastname:  lastname,
		Password:  string(hashedPassword),
		Role:      "user", // default role is "user"
	}

	_, err = s.UserCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		return nil, "", errors.New("failed to create user")
	}

	// Generate a JWT for the newly created user
	token, err := utils.GenerateJWT(newUser.ID.Hex(), newUser.Role) // using the GenerateJWT function
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &newUser, token, nil
}

// Login handles user login to dashboard
func (s *AuthService) Login(email, password string) (string, error) {
	var user model.User
	err := s.UserCollection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println("Error comparing password:", err)
		return "", errors.New("invalid email or password")
	}

	// Generate JWT token with user ID and role
	token, err := utils.GenerateJWT(user.ID.Hex(), user.Role) // using the GenerateJWT function
	if err != nil {
		log.Println("Error generating JWT:", err)
		return "", errors.New("failed to generate token")
	}

	return token, nil
}
