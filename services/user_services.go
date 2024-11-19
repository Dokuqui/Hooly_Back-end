package services

import (
	"context"
	"errors"
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	UserCollection *mongo.Collection
}

func NewUserService() *UserService {
	return &UserService{
		UserCollection: db.GetCollection("user"),
	}
}

// CreateUser handles the creation of a new user
func (s *UserService) CreateUser(firstName, lastName, email, password string, role model.Role) (*model.User, error) {
	var existingUser model.User
	err := s.UserCollection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&existingUser)
	if err == nil {
		// If user already exists, return an error
		return nil, errors.New("email already registered")
	}

	// Hash password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create the new user
	newUser := model.User{
		ID:        primitive.NewObjectID(),
		Firstname: firstName,
		Lastname:  lastName,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
	}

	// Insert the new user into the collection
	_, err = s.UserCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	return &newUser, nil
}

// GetAllUsers fetches all users from the collection
func (s *UserService) GetAllUsers() ([]model.User, error) {
	var users []model.User

	// Fetch all users from the collection
	cursor, err := s.UserCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, errors.New("failed to retrieve users")
	}
	defer cursor.Close(context.TODO())

	// Iterate over the cursor and append each user to the users slice
	for cursor.Next(context.TODO()) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			return nil, errors.New("failed to decode user")
		}
		users = append(users, user)
	}

	// Check for any errors encountered during iteration
	if err := cursor.Err(); err != nil {
		return nil, errors.New("cursor error")
	}

	return users, nil
}

// GetUserById fetches a user by their ID
func (s *UserService) GetUserById(userID string) (*model.User, error) {
	var user model.User
	objectId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Fetch the user from collection
	err = s.UserCollection.FindOne(context.TODO(), bson.M{"_id": objectId}).Decode(&user)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// UpdateUser allows an admin to update any user's details or a user to update their own
func (s *UserService) UpdateUser(userID string, updatedUser model.User) (*model.User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Check if the user exists
	var user model.User
	err = s.UserCollection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Update the user's fields
	update := bson.M{}
	if updatedUser.Firstname != "" {
		update["firstname"] = updatedUser.Firstname
	}
	if updatedUser.Lastname != "" {
		update["lastname"] = updatedUser.Lastname
	}
	if updatedUser.Email != "" {
		update["email"] = updatedUser.Email
	}
	if updatedUser.Password != "" {
		// Hash the password if it's being updated
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedUser.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("failed to hash new password")
		}
		update["password"] = string(hashedPassword)
	}

	// Update the user in the collection
	_, err = s.UserCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": objectID},
		bson.M{"$set": update},
	)
	if err != nil {
		return nil, errors.New("failed to update user")
	}

	// Fetch the updated user
	err = s.UserCollection.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser deletes a user from the collection (admin only)
func (s *UserService) DeleteUser(userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// Delete the user from the collection
	_, err = s.UserCollection.DeleteOne(context.TODO(), bson.M{"_id": objectID})
	if err != nil {
		return errors.New("failed to delete user")
	}

	return nil
}
