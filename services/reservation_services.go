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
	// Validation: Ensure Spot is available and FoodTruck hasn't booked in the same week.
	existingFilter := bson.M{
		"spot_id":       reservation.SpotID,
		"date":          reservation.Date,
		"food_truck_id": reservation.FoodTruckID,
	}
	count, err := s.ReservationCollection.CountDocuments(ctx, existingFilter)
	if err != nil {
		return errors.New("failed to check existing reservations")
	}
	if count > 0 {
		return errors.New("spot is already reserved for this date")
	}

	// Insert reservation
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

	update := bson.M{"$set": updateData}
	_, err := s.ReservationCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("failed to update reservation or not authorized")
	}
	return nil
}

// DeleteReservation deletes a reservation by ID, optionally scoped by user ID.
func (s *ReservationService) DeleteReservation(ctx context.Context, reservationID primitive.ObjectID, userID primitive.ObjectID) error {
	filter := bson.M{"_id": reservationID}
	if !userID.IsZero() {
		filter["user_id"] = userID
	}

	_, err := s.ReservationCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete reservation or not authorized")
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
