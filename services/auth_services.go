package services

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var jwtSecret = []byte("secret")

type AuthService struct {
	UserCollection *mongo.Collection
}

func NewAuthService() *AuthService {
	return &AuthService{
		UserCollection: db.GetCollection("user"),
	}
}

// Signup handles new user creation
func (s *AuthService) Signup(email, password string) (*model.User, error) {
	var existingUser model.User
	err := s.UserCollection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&existingUser)
	if err == nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	newUser := model.User{
		ID:       primitive.NewObjectID(),
		Email:    email,
		Password: string(hashedPassword),
	}

	_, err = s.UserCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	return &newUser, nil
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
		return "", errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := generateJWT(user.ID.Hex())
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}

// generateJWT generating JWT token for login
func generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
