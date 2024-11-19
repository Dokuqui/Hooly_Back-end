package services

import (
	"errors"
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/model"
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
		ParkingSpotCollection: db.GetCollection("parkingspots"),
	}
}

// Get all spots
func (s *ParkingSpotService) ListAllParkingSpots(dayOfWeek string) ([]model.ParkingSpot, error) {
	filter := bson.M{}
	if dayOfWeek != "" {
		filter["day_of_week"] = dayOfWeek
	}

	cursor, err := s.ParkingSpotCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var spots []model.ParkingSpot
	for cursor.Next(context.TODO()) {
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
