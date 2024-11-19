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

// Find by ID
func (s *FoodtruckService) FindFoodtruckByID(id string) (*model.Foodtruck, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	var foodtruck model.Foodtruck
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = s.FoodtruckCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&foodtruck)
	if err != nil {
		return nil, err
	}

	return &foodtruck, nil
}

// Find by NAME
func (s *FoodtruckService) FindFoodtruckByName(name string) ([]model.Foodtruck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.FoodtruckCollection.Find(ctx, bson.M{"name": name})
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

// Add a foodtruck
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

// Update a foodtruck
func (s *FoodtruckService) UpdateFoodtruck(id string, foodtruck *model.Foodtruck) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid ID format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": foodtruck}

	_, err = s.FoodtruckCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// Delete a foodtruck
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
