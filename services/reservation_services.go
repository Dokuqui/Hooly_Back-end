package services

import (
	"context"
	"errors"
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ReservationService struct {
	ReservationCollection *mongo.Collection
	ParkingSpotCollection *mongo.Collection
	UserCollection        *mongo.Collection
}

func NewReservationService() *ReservationService {
	return &ReservationService{
		ReservationCollection: db.GetCollection("reservation"),
		ParkingSpotCollection: db.GetCollection("parkingSpot"),
		UserCollection:        db.GetCollection("user"),
	}
}

// GetAllReservations retrieves all reservations (admin use case).
func (s *ReservationService) GetAllReservations(ctx context.Context) ([]model.Reservation, error) {
	cursor, err := s.ReservationCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var reservations []model.Reservation
	if err = cursor.All(ctx, &reservations); err != nil {
		return nil, err
	}

	return reservations, nil
}

// GetUserReservations retrieves all reservations associated with a specific user.
func (s *ReservationService) GetUserReservations(ctx context.Context, userID primitive.ObjectID) ([]model.Reservation, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := s.ReservationCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var reservations []model.Reservation
	if err = cursor.All(ctx, &reservations); err != nil {
		return nil, err
	}

	return reservations, nil
}

// GetReservationByID retrieves a reservation by ID, optionally scoped by user ID from context.
func (s *ReservationService) GetReservationByID(ctx context.Context, reservationID primitive.ObjectID, userID primitive.ObjectID) (*model.Reservation, error) {
	filter := bson.M{"_id": reservationID}
	if !userID.IsZero() {
		filter["user_id"] = userID
	}

	var reservation model.Reservation
	err := s.ReservationCollection.FindOne(ctx, filter).Decode(&reservation)
	if err != nil {
		return nil, err
	}

	return &reservation, nil
}

// CreateReservation creates a new reservation.
func (s *ReservationService) CreateReservation(ctx context.Context, reservation *model.Reservation) error {
	// Validate the reservation date: it should be in the future and not today
	if reservation.Date.Before(time.Now().Add(time.Hour * 24)) {
		return errors.New("cannot reserve a spot for a past date or today")
	}

	// Ensure the food truck has not reserved a spot for the same week
	// Checking the reservations within the past 7 days
	existingFilter := bson.M{
		"food_truck_id": reservation.FoodTruckID,
		"date": bson.M{
			"$gte": time.Now().Add(-7 * 24 * time.Hour), // Look for reservations within the past 7 days
		},
	}
	count, err := s.ReservationCollection.CountDocuments(ctx, existingFilter)
	if err != nil {
		return errors.New("failed to check existing reservations")
	}
	if count > 0 {
		return errors.New("food truck already has a reservation for this week")
	}

	// Ensure the parking spot is available for the given day
	parkingSpotFilter := bson.M{"_id": reservation.SpotID, "status": "available"}
	parkingSpot := model.ParkingSpot{}
	err = s.ParkingSpotCollection.FindOne(ctx, parkingSpotFilter).Decode(&parkingSpot)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return errors.New("spot is not available")
		}
		return err
	}

	// Ensure the max capacity is not exceeded
	// Here you should check the number of existing reservations for the specific day
	spotCount, err := s.ReservationCollection.CountDocuments(ctx, bson.M{"spot_id": reservation.SpotID, "date": reservation.Date})
	if err != nil {
		return errors.New("failed to check spot reservations")
	}

	// Cast MaxCapacity to int64 for comparison
	if spotCount >= int64(parkingSpot.MaxCapacity) {
		return errors.New("no available spots for this day")
	}

	// Mark the parking spot as reserved
	updateSpot := bson.M{"$set": bson.M{"status": "reserved"}}
	_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, updateSpot)
	if err != nil {
		return errors.New("failed to update parking spot status")
	}

	// Insert the reservation into the reservation collection
	reservation.CreatedAt = time.Now()
	_, err = s.ReservationCollection.InsertOne(ctx, reservation)
	return err
}

// UpdateReservation updates an existing reservation.
func (s *ReservationService) UpdateReservation(ctx context.Context, reservationID primitive.ObjectID, updateData bson.M, userID primitive.ObjectID) error {
	filter := bson.M{"_id": reservationID}
	if !userID.IsZero() {
		filter["user_id"] = userID
	}

	// Check if the reservation spot will change
	var reservation model.Reservation
	err := s.ReservationCollection.FindOne(ctx, filter).Decode(&reservation)
	if err != nil {
		return errors.New("reservation not found")
	}

	// If spot is changing, release the old spot
	if updateData["spot_id"] != nil && updateData["spot_id"] != reservation.SpotID {
		// Release the old spot
		releaseSpot := bson.M{"$set": bson.M{"status": "available"}}
		_, err := s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, releaseSpot)
		if err != nil {
			return err
		}

		// Mark the new spot as reserved
		newSpotID := updateData["spot_id"].(primitive.ObjectID)
		updateSpot := bson.M{"$set": bson.M{"status": "reserved"}}
		_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": newSpotID}, updateSpot)
		if err != nil {
			return errors.New("failed to update parking spot status")
		}
	}

	// Update reservation
	update := bson.M{"$set": updateData}
	_, err = s.ReservationCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("failed to update reservation")
	}
	return nil
}

// DeleteReservation releases the parking spot when the reservation is deleted.
func (s *ReservationService) DeleteReservation(ctx context.Context, reservationID primitive.ObjectID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": reservationID}
	if !userID.IsZero() {
		filter["user_id"] = userID
	}

	var reservation model.Reservation
	err := s.ReservationCollection.FindOne(ctx, filter).Decode(&reservation)
	if err != nil {
		return errors.New("reservation not found")
	}

	// Release the parking spot
	releaseSpot := bson.M{"$set": bson.M{"status": "available"}}
	_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, releaseSpot)
	if err != nil {
		return errors.New("failed to release parking spot")
	}

	// Delete the reservation
	_, err = s.ReservationCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete reservation")
	}

	return nil
}

// AdminDeleteReservation deletes a reservation without user_id restrictions (admin functionality).
func (s *ReservationService) AdminDeleteReservation(ctx context.Context, reservationID primitive.ObjectID) error {
	filter := bson.M{"_id": reservationID}

	_, err := s.ReservationCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete reservation")
	}

	return nil
}
