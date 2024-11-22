package services

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

// MockMongoCollection is a mock implementation of the mongo.Collection type
type MockMongoCollection struct {
	mock.Mock
}

func (m *MockMongoCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockMongoCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

// Mocked AuthService depends on a Collection type
type MockAuthService struct {
	UserCollection *MockMongoCollection
}

// Signup method of AuthService
func (a *MockAuthService) Signup(email, firstname, lastname, password string) (*mongo.SingleResult, string, error) {
	// Mock signup process logic (simplified)
	filter := bson.M{"email": email}
	result := a.UserCollection.FindOne(context.Background(), filter)

	// If email is already in the database, return an error
	if result != nil && result.Err() == nil {
		return nil, "", errors.New("email already registered")
	}

	// Insert new user logic (not actually implemented here for simplicity)
	_, err := a.UserCollection.InsertOne(context.Background(), bson.M{
		"email":     email,
		"firstname": firstname,
		"lastname":  lastname,
		"password":  password,
	})
	if err != nil {
		return nil, "", err
	}

	// Simulating successful registration and generating a fake token (simplified)
	return result, "fake-jwt-token", nil
}

func TestSignup(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		firstname    string
		lastname     string
		password     string
		mockBehavior func(mockUserCollection *MockMongoCollection)
		expectedErr  error
	}{
		{
			name:      "Successful Signup",
			email:     "newuser@example.com",
			firstname: "John",
			lastname:  "Doe",
			password:  "securepassword123",
			mockBehavior: func(mockUserCollection *MockMongoCollection) {
				// Mock successful FindOne response (simulate no existing user)
				mockUserCollection.On("FindOne", mock.Anything, mock.MatchedBy(func(filter interface{}) bool {
					bsonFilter, ok := filter.(bson.M)
					return ok && bsonFilter["email"] == "newuser@example.com"
				})).Return(&mongo.SingleResult{}, nil)

				// Mock InsertOne response (simulate successful user insertion)
				mockUserCollection.On("InsertOne", mock.Anything, mock.Anything).Return(&mongo.InsertOneResult{InsertedID: primitive.NewObjectID()}, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mock user collection
			mockUserCollection := new(MockMongoCollection)

			// Set the mock behavior
			tt.mockBehavior(mockUserCollection)

			// Create the AuthService with the mocked UserCollection
			authService := &MockAuthService{
				UserCollection: mockUserCollection,
			}

			// Run the Signup method
			_, _, err := authService.Signup(tt.email, tt.firstname, tt.lastname, tt.password)

			// Check the error
			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Fatalf("expected error %v, but got %v", tt.expectedErr, err)
			}
		})
	}
}
