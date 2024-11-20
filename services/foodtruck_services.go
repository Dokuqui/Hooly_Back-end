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

// GetAllFoodTrucks retrieves all food trucks (admin use case).
func (s *FoodtruckService) GetAllFoodTrucks(ctx context.Context) ([]model.Foodtruck, error) {
	cursor, err := s.FoodtruckCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var foodtrucks []model.Foodtruck
	if err = cursor.All(ctx, &foodtrucks); err != nil {
		return nil, err
	}

	return foodtrucks, nil
}

// GetUserFoodTrucks retrieves all food trucks for a specific user
func (s *FoodtruckService) GetUserFoodTrucks(ctx context.Context, userID primitive.ObjectID) ([]model.Foodtruck, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := s.FoodtruckCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var foodtrucks []model.Foodtruck
	if err = cursor.All(ctx, &foodtrucks); err != nil {
		return nil, err
	}

	return foodtrucks, nil
}

// GetFoodTruckByID retrieves a reservation by ID, optionally scoped by user ID from context.
func (s *FoodtruckService) GetFoodTruckByID(ctx context.Context, foodtruckId primitive.ObjectID, userID primitive.ObjectID) (*model.Foodtruck, error) {
	filter := bson.M{"_id": foodtruckId}
	if !userID.IsZero() {
		filter["user_id"] = userID
	}

	var foodtruck model.Foodtruck
	err := s.FoodtruckCollection.FindOne(ctx, filter).Decode(&foodtruck)
	if err != nil {
		return nil, err
	}

	return &foodtruck, nil
}

// AddFoodtruck Add a foodtruck
func (s *FoodtruckService) AddFoodtruck(foodtruck *model.Foodtruck) (*model.Foodtruck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Assign a unique ID to the food truck
	foodtruck.ID = primitive.NewObjectID()

	// Insert the food truck into the MongoDB collection
	_, err := s.FoodtruckCollection.InsertOne(ctx, foodtruck)
	if err != nil {
		return nil, err
	}

	return foodtruck, nil
}

// UpdateFoodtruck Update a foodtruck
func (s *FoodtruckService) UpdateFoodtruck(ctx context.Context, foodTruckID primitive.ObjectID, updateData bson.M, userID primitive.ObjectID) error {
	filter := bson.M{"_id": foodTruckID}
	if !userID.IsZero() {
		filter["user_id"] = userID
	}

	update := bson.M{"$set": updateData}
	_, err := s.FoodtruckCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("failed to update foodtruck or not authorized")
	}
	return nil
}

// DeleteFoodtruck deletes a reservation by ID, optionally scoped by user ID.
func (s *FoodtruckService) DeleteFoodtruck(ctx context.Context, foodtruckID primitive.ObjectID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": foodtruckID}
	if !userID.IsZero() {
		filter["user_id"] = userID
	}

	_, err := s.FoodtruckCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete reservation or not authorized")
	}

	return nil
}

// AdminDeleteFoodtruck deletes a reservation without user_id restrictions (admin functionality).
func (s *FoodtruckService) AdminDeleteFoodtruck(ctx context.Context, foodtruckID primitive.ObjectID) error {
	filter := bson.M{"_id": foodtruckID}

	_, err := s.FoodtruckCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete reservation")
	}

	return nil
}
