package services

import (
	"errors"
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
	"time"
)

// FoodtruckService provides CRUD operations for Foodtruck
type FoodtruckService struct {
	FoodtruckCollection *mongo.Collection
}

// NewFoodtruckService creates a new FoodtruckService
func NewFoodtruckService() *FoodtruckService {
	return &FoodtruckService{
		FoodtruckCollection: db.GetCollection("foodtruck"),
	}
}

// FindFoodtruckByID Find food truck by id
func (s *FoodtruckService) FindFoodtruckByID(userID, id string) (*model.Foodtruck, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	var foodtruck model.Foodtruck
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = s.FoodtruckCollection.FindOne(ctx, bson.M{
		"_id":     objectID,
		"user_id": userID,
	}).Decode(&foodtruck)
	if err != nil {
		return nil, err
	}

	return &foodtruck, nil
}

// FindFoodtruckByName Find by NAME
func (s *FoodtruckService) FindFoodtruckByName(name, userID string) ([]model.Foodtruck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.FoodtruckCollection.Find(ctx, bson.M{
		"name":    name,
		"user_id": userID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var foodtrucks []model.Foodtruck
	if err := cursor.All(ctx, &foodtrucks); err != nil {
		return nil, err
	}

	return foodtrucks, nil
}

// GetAllFoodtrucksByUserID retrieves all food trucks for a specific user
func (s *FoodtruckService) GetAllFoodtrucksByUserID(userID string) ([]model.Foodtruck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use the userID as a filter
	cursor, err := s.FoodtruckCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the retrieved documents into a slice of Foodtruck
	var foodtrucks []model.Foodtruck
	if err := cursor.All(ctx, &foodtrucks); err != nil {
		return nil, err
	}

	return foodtrucks, nil
}

// AddFoodtruck Add a foodtruck
func (s *FoodtruckService) AddFoodtruck(foodtruck *model.Foodtruck) (*model.Foodtruck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	foodtruck.ID = primitive.NewObjectID()
	_, err := s.FoodtruckCollection.InsertOne(ctx, foodtruck)
	if err != nil {
		return nil, err
	}

	return foodtruck, nil
}

// UpdateFoodtruck Update a foodtruck
func (s *FoodtruckService) UpdateFoodtruck(id string, userID primitive.ObjectID, foodtruck *model.Foodtruck) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": objectID, "user_id": userID}
	update := bson.M{"$set": foodtruck}

	_, err = s.FoodtruckCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// DeleteFoodtruck Delete a foodtruck
func (s *FoodtruckService) DeleteFoodtruck(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = s.FoodtruckCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	return nil
}
