package services

import (
	"context"
	"errors"
	"fmt"
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

	// Enrich each reservation with spot details
	for i, reservation := range reservations {
		// Fetch the parking spot associated with the reservation
		parkingSpotFilter := bson.M{"_id": reservation.SpotID}
		var parkingSpot model.ParkingSpot
		err := s.ParkingSpotCollection.FindOne(ctx, parkingSpotFilter).Decode(&parkingSpot)
		if err != nil {
			return nil, err
		}

		reservations[i].SpotNumber = reservation.SpotNumber // This is already part of reservation
	}

	return reservations, nil
}

// GetAllUserReservations retrieves all reservations for a user (without userID and foodTruckID).
func (s *ReservationService) GetAllUserReservations(ctx context.Context) ([]model.Reservation, error) {
	cursor, err := s.ReservationCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reservations []model.Reservation
	if err = cursor.All(ctx, &reservations); err != nil {
		return nil, err
	}

	// Enrich each reservation with spot details and remove userID and foodTruckID for user
	for i, reservation := range reservations {
		// Fetch the parking spot associated with the reservation
		parkingSpotFilter := bson.M{"_id": reservation.SpotID}
		var parkingSpot model.ParkingSpot
		err := s.ParkingSpotCollection.FindOne(ctx, parkingSpotFilter).Decode(&parkingSpot)
		if err != nil {
			return nil, err
		}

		// Remove sensitive fields for user view
		reservations[i].UserID = primitive.NilObjectID
		reservations[i].FoodTruckID = primitive.NilObjectID

		reservations[i].SpotNumber = reservation.SpotNumber
	}

	return reservations, nil
}

// GetUserReservations retrieves all reservations associated with a specific user, including spot information.
func (s *ReservationService) GetUserReservations(ctx context.Context, userID primitive.ObjectID) ([]model.Reservation, error) {
	filter := bson.M{"user_id": userID}

	// Find all reservations for the user
	cursor, err := s.ReservationCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var reservations []model.Reservation
	if err = cursor.All(ctx, &reservations); err != nil {
		return nil, err
	}

	// Enrich each reservation with spot details
	for i, reservation := range reservations {
		// Fetch the parking spot associated with the reservation
		parkingSpotFilter := bson.M{"_id": reservation.SpotID}
		var parkingSpot model.ParkingSpot
		err := s.ParkingSpotCollection.FindOne(ctx, parkingSpotFilter).Decode(&parkingSpot)
		if err != nil {
			return nil, err
		}

		reservations[i].SpotNumber = reservation.SpotNumber // This is already part of reservation
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

// CreateReservation creates a new reservation with the ability to choose the spot number.
func (s *ReservationService) CreateReservation(ctx context.Context, reservation *model.Reservation) error {
	// Validate the reservation date: it should be in the future and not today
	if reservation.Date.Before(time.Now().Add(time.Hour * 24)) {
		return errors.New("cannot reserve a spot for a past date or today")
	}

	// Ensure the food truck has not reserved a spot for the same week
	existingFilter := bson.M{
		"food_truck_id": reservation.FoodTruckID,
		"date": bson.M{
			"$gte": time.Now().Add(-7 * 24 * time.Hour),
		},
	}

	count, err := s.ReservationCollection.CountDocuments(ctx, existingFilter)
	if err != nil {
		return errors.New("failed to check existing reservations")
	}
	if count > 0 {
		return errors.New("food truck already has a reservation for this week")
	}

	// Ensure the SpotID is in ObjectID format
	spotID, err := primitive.ObjectIDFromHex(reservation.SpotID.Hex())
	if err != nil {
		return errors.New("invalid SpotID")
	}

	// Ensure the parking spot is available for the given day and spot number
	parkingSpotFilter := bson.M{"_id": spotID}
	parkingSpot := model.ParkingSpot{}
	err = s.ParkingSpotCollection.FindOne(ctx, parkingSpotFilter).Decode(&parkingSpot)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			return errors.New("spot is not available")
		}
		return err
	}

	// Ensure the chosen spot number is available (i.e., not already reserved)
	spotNumberAvailable := false
	for _, num := range parkingSpot.SpotNumbers {
		if num == reservation.SpotNumber {
			spotNumberAvailable = true
			break
		}
	}
	if !spotNumberAvailable {
		return fmt.Errorf("spot number %d is not available for reservation", reservation.SpotNumber)
	}

	// Ensure the max capacity is not exceeded (calculate used spots for the day)
	spotCount, err := s.ReservationCollection.CountDocuments(ctx, bson.M{
		"spot_id": spotID,
		"date":    reservation.Date,
	})
	if err != nil {
		return errors.New("failed to check spot reservations")
	}

	// Decrement the available capacity (convert spotCount to int for comparison)
	if int(spotCount) >= parkingSpot.MaxCapacity {
		return errors.New("no available spots for this day")
	}

	// Insert the reservation into the reservation collection
	reservation.CreatedAt = time.Now()
	result, err := s.ReservationCollection.InsertOne(ctx, reservation)
	if err != nil {
		return err
	}

	// Ensure the reservation ID is populated
	reservation.ID = result.InsertedID.(primitive.ObjectID)

	// Update the parking spot's reserved count and mark the spot as reserved
	update := bson.M{
		"$inc": bson.M{
			"reserved_count": 1,
		},
		"$push": bson.M{
			"reserved_spots": reservation.SpotNumber, // Track which spots are reserved for the given day
		},
	}

	// Update the parking spot with the new reservation details
	_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": spotID}, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *ReservationService) UpdateReservation(ctx context.Context, reservationID primitive.ObjectID, updateData bson.M, userID primitive.ObjectID) error {
	// Only filter by reservation ID
	filter := bson.M{"_id": reservationID}

	var reservation model.Reservation
	err := s.ReservationCollection.FindOne(ctx, filter).Decode(&reservation)
	if err != nil {
		return errors.New("reservation not found")
	}

	// If spot number is changing, update the parking spots accordingly
	if updateData["spot_number"] != nil && updateData["spot_number"] != reservation.SpotNumber {
		// Release the old spot
		updateOldSpot := bson.M{"$pull": bson.M{"reserved_spots": reservation.SpotNumber}, "$inc": bson.M{"reserved_count": -1}}
		_, err := s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, updateOldSpot)
		if err != nil {
			return err
		}

		// Update the reservation with the new spot number
		newSpotNumber := updateData["spot_number"].(int)
		updateData["spot_number"] = newSpotNumber // Ensure reservation's spot number is updated

		// Add the new spot number to the reserved spots
		updateNewSpot := bson.M{"$push": bson.M{"reserved_spots": newSpotNumber}, "$inc": bson.M{"reserved_count": 1}}
		_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, updateNewSpot)
		if err != nil {
			return errors.New("failed to update parking spot status")
		}
	}

	// Update reservation with new data
	update := bson.M{"$set": updateData}
	_, err = s.ReservationCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("failed to update reservation")
	}

	return nil
}

func (s *ReservationService) DeleteReservation(ctx context.Context, reservationID primitive.ObjectID, userID primitive.ObjectID) error {
	// Find the reservation by ID (and optionally user ID if provided)
	filter := bson.M{"_id": reservationID}
	if !userID.IsZero() {
		filter["user_id"] = userID
	}

	var reservation model.Reservation
	err := s.ReservationCollection.FindOne(ctx, filter).Decode(&reservation)
	if err != nil {
		return errors.New("reservation not found")
	}

	// Find the associated parking spot
	parkingSpot := model.ParkingSpot{}
	err = s.ParkingSpotCollection.FindOne(ctx, bson.M{"_id": reservation.SpotID}).Decode(&parkingSpot)
	if err != nil {
		return errors.New("parking spot not found")
	}

	// Decrease the reserved count for the parking spot
	updateSpot := bson.M{"$pull": bson.M{"reserved_spots": reservation.SpotNumber}, "$inc": bson.M{"reserved_count": -1}} // Decrement reserved count and remove spot number
	_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, updateSpot)
	if err != nil {
		return errors.New("failed to update parking spot capacity")
	}

	// Delete the reservation
	_, err = s.ReservationCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete reservation")
	}

	return nil
}

func (s *ReservationService) AdminDeleteReservation(ctx context.Context, reservationID primitive.ObjectID) error {
	// Find the reservation by ID
	filter := bson.M{"_id": reservationID}
	var reservation model.Reservation
	err := s.ReservationCollection.FindOne(ctx, filter).Decode(&reservation)
	if err != nil {
		return errors.New("reservation not found")
	}

	// Find the associated parking spot
	parkingSpot := model.ParkingSpot{}
	err = s.ParkingSpotCollection.FindOne(ctx, bson.M{"_id": reservation.SpotID}).Decode(&parkingSpot)
	if err != nil {
		return errors.New("parking spot not found")
	}

	// Decrease the reserved count for the parking spot
	updateSpot := bson.M{"$pull": bson.M{"reserved_spots": reservation.SpotNumber}, "$inc": bson.M{"reserved_count": -1}}
	_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, updateSpot)
	if err != nil {
		return errors.New("failed to update parking spot capacity")
	}

	// Delete the reservation
	_, err = s.ReservationCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete reservation")
	}

	return nil
}
