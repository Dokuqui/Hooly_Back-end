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
	"log"
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
		Role:      "user",
	}

	_, err = s.UserCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		return nil, "", errors.New("failed to create user")
	}

	// Generate a JWT for the newly created user
	token, err := generateJWT(newUser.ID.Hex(), string(newUser.Role))
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return &newUser, token, nil
}

// Login handles user login to dashboard
func (s *AuthService) Login(email, password string) (string, *model.User, error) {
	var user model.User
	err := s.UserCollection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		log.Println("Error finding user:", err)
		return "", nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println("Error comparing password:", err)
		return "", nil, errors.New("invalid email or password")
	}

	// Generate JWT token with user ID and role
	token, err := generateJWT(user.ID.Hex(), string(user.Role))
	if err != nil {
		log.Println("Error generating JWT:", err)
		return "", nil, errors.New("failed to generate token")
	}

	// Exclude the password from the returned user object for security
	user.Password = ""
	return token, &user, nil
}

// generateJWT generating JWT token for login
func generateJWT(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Println("Error signing JWT:", err)
		return "", err
	}

	return signedToken, nil
}
