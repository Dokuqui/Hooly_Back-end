package services

import (
	"errors"
	"fmt"
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/model"
	"gitlab.com/hooly2/back/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type ParkingSpotService struct {
	ParkingSpotCollection *mongo.Collection
}

func NewParkingSpotService() *ParkingSpotService {
	return &ParkingSpotService{
		ParkingSpotCollection: db.GetCollection("parkingSpot"),
	}
}

// CreateParkingSpot Create parking spot for a specific day of the week
func (s *ParkingSpotService) CreateParkingSpot(dayOfWeek string, ctx context.Context) (*model.ParkingSpot, error) {
	// Ensure that only valid days are used
	if !utils.IsValidDayOfWeek(dayOfWeek) {
		return nil, errors.New("invalid day of the week")
	}

	// Check if the parking spot already exists for the given day
	var existingSpot model.ParkingSpot
	err := s.ParkingSpotCollection.FindOne(ctx, bson.M{"day_of_week": dayOfWeek}).Decode(&existingSpot)
	if err == nil {
		return nil, errors.New("parking spot already exists for this day")
	}

	// Determine the number of available parking spots for the day
	totalSpaces := 7
	if dayOfWeek == "Friday" {
		totalSpaces = 6
	}

	// Create the new parking spot document
	newSpot := model.ParkingSpot{
		ID:          primitive.NewObjectID(),
		Day:         dayOfWeek,
		MaxCapacity: totalSpaces,
		Reserved:    false,
	}

	// Insert the new parking spot into the collection
	_, err = s.ParkingSpotCollection.InsertOne(ctx, newSpot)
	if err != nil {
		return nil, fmt.Errorf("failed to create parking spot: %v", err)
	}

	return &newSpot, nil
}

// ListAllParkingSpots Get all spots
func (s *ParkingSpotService) ListAllParkingSpots(dayOfWeek string, ctx context.Context) ([]model.ParkingSpot, error) {
	filter := bson.M{}
	if dayOfWeek != "" {
		filter["day_of_week"] = dayOfWeek
	}

	cursor, err := s.ParkingSpotCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var spots []model.ParkingSpot
	for cursor.Next(ctx) {
		var spot model.ParkingSpot
		if err := cursor.Decode(&spot); err != nil {
			return nil, err
		}
		spots = append(spots, spot)
	}

	return spots, nil
}

// Check spot availability
func (s *ParkingSpotService) IsSpotAvailable(spotID primitive.ObjectID) (bool, error) {
	var spot model.ParkingSpot
	err := s.ParkingSpotCollection.FindOne(context.TODO(), bson.M{"_id": spotID}).Decode(&spot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("parking spot does not exist")
		}
		return false, err
	}

	return !spot.Reserved, nil
}

// Update reservation status
func (s *ParkingSpotService) UpdateReservationStatus(spotID primitive.ObjectID, reserved bool) error {
	_, err := s.ParkingSpotCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": spotID},
		bson.M{"$set": bson.M{"reserved": reserved}},
	)
	return err
}
