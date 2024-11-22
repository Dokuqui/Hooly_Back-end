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
	// Validate the day of the week
	if !utils.IsValidDayOfWeek(dayOfWeek) {
		return nil, errors.New("invalid day of the week")
	}

	// Check if the parking spot already exists
	var existingSpot model.ParkingSpot
	err := s.ParkingSpotCollection.FindOne(ctx, bson.M{"day_of_week": dayOfWeek}).Decode(&existingSpot)
	if err == nil {
		return nil, errors.New("parking spot already exists for this day")
	} else if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to query parking spot: %v", err)
	}

	// Determine the total number of spots based on the day
	var totalSpaces int
	var spotNumbers []int
	if dayOfWeek == "Friday" {
		totalSpaces = 6
		spotNumbers = []int{1, 2, 3, 4, 5, 6} // Friday has 6 spots
	} else {
		totalSpaces = 7
		spotNumbers = []int{1, 2, 3, 4, 5, 6, 7} // Other days have 7 spots
	}

	// Create the new parking spot document
	newSpot := model.ParkingSpot{
		ID:            primitive.NewObjectID(),
		Day:           dayOfWeek,
		MaxCapacity:   totalSpaces,
		ReservedCount: 0,
		SpotNumbers:   spotNumbers,
		ReservedSpots: []int{},
	}

	// Insert the new parking spot
	_, err = s.ParkingSpotCollection.InsertOne(ctx, newSpot)
	if err != nil {
		return nil, fmt.Errorf("failed to create parking spot: %v", err)
	}

	return &newSpot, nil
}

// ListAllParkingSpots retrieves all parking spots, filtered by day if specified
func (s *ParkingSpotService) ListAllParkingSpots(dayOfWeek string, ctx context.Context) ([]model.ParkingSpot, error) {
	filter := bson.M{}
	if dayOfWeek != "" {
		filter["day_of_week"] = dayOfWeek
	}

	// Query the collection to fetch parking spots
	cursor, err := s.ParkingSpotCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch parking spots: %v", err)
	}
	defer cursor.Close(ctx)

	var spots []model.ParkingSpot
	for cursor.Next(ctx) {
		var spot model.ParkingSpot
		if err := cursor.Decode(&spot); err != nil {
			return nil, fmt.Errorf("failed to decode parking spot: %v", err)
		}
		spots = append(spots, spot)
	}

	return spots, nil
}

// UpdateReservationStatus updates the reservation status of a parking spot
func (s *ParkingSpotService) UpdateReservationStatus(spotID primitive.ObjectID, reserved bool, ctx context.Context) error {
	update := bson.M{"$set": bson.M{"reserved": reserved}}
	_, err := s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": spotID}, update)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %v", err)
	}
	return nil
}
