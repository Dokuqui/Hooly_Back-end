package services

import (
	"context"
	"errors"
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
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
		log.Println("Error converting SpotID to ObjectID:", err)
		return errors.New("invalid SpotID")
	}

	// Ensure the parking spot is available for the given day
	parkingSpotFilter := bson.M{"_id": spotID}
	parkingSpot := model.ParkingSpot{}
	err = s.ParkingSpotCollection.FindOne(ctx, parkingSpotFilter).Decode(&parkingSpot)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			log.Println("Spot not found or not available:", err)
			return errors.New("spot is not available")
		}
		log.Println("Error retrieving parking spot:", err)
		return err
	}

	log.Printf("Found parking spot: %+v\n", parkingSpot)

	// Ensure the max capacity is not exceeded (calculate used spots for the day)
	spotCount, err := s.ReservationCollection.CountDocuments(ctx, bson.M{"spot_id": spotID, "date": reservation.Date})
	if err != nil {
		log.Println("Error checking reservation count:", err)
		return errors.New("failed to check spot reservations")
	}

	log.Printf("Current reservation count for SpotID %s on %s: %d\n", spotID.Hex(), reservation.Date.Weekday(), spotCount)

	// Decrement the available capacity (convert spotCount to int for comparison)
	if int(spotCount) >= parkingSpot.MaxCapacity {
		log.Printf("No available spots for SpotID: %s, MaxCapacity: %d, spotCount: %d\n", spotID.Hex(), parkingSpot.MaxCapacity, spotCount)
		return errors.New("no available spots for this day")
	}

	// Insert the reservation into the reservation collection
	reservation.CreatedAt = time.Now()
	result, err := s.ReservationCollection.InsertOne(ctx, reservation)
	if err != nil {
		log.Println("Error inserting reservation:", err)
		return err
	}

	// Ensure the reservation ID is populated
	reservation.ID = result.InsertedID.(primitive.ObjectID)

	// Increment the reserved count for the parking spot
	update := bson.M{
		"$inc": bson.M{
			"reserved_count": 1, // Increment the reserved spots count
		},
	}

	// Update the parking spot's reserved count
	_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": spotID}, update)
	if err != nil {
		log.Println("Error updating reserved count:", err)
		return err
	}

	// Check if the parking spot is now fully reserved and update its status
	if parkingSpot.ReservedCount+1 >= parkingSpot.MaxCapacity {
		log.Printf("Spot %s marked as reserved for the day\n", spotID.Hex())
	} else {
		log.Printf("Reserved count for SpotID %s updated: %d\n", spotID.Hex(), parkingSpot.ReservedCount+1)
	}

	log.Printf("Reservation created successfully for SpotID: %s, FoodTruckID: %s\n", spotID.Hex(), reservation.FoodTruckID.Hex())
	return nil
}

// UpdateReservation updates an existing reservation.
func (s *ReservationService) UpdateReservation(ctx context.Context, reservationID primitive.ObjectID, updateData bson.M, userID primitive.ObjectID) error {
	// Only filter by reservation ID
	filter := bson.M{"_id": reservationID}
	log.Println("Filter used for querying reservation:", filter)

	var reservation model.Reservation
	err := s.ReservationCollection.FindOne(ctx, filter).Decode(&reservation)
	if err != nil {
		log.Println("Error finding reservation:", err)
		return errors.New("reservation not found")
	}

	// If spot is changing, release the old spot
	if updateData["spot_id"] != nil && updateData["spot_id"] != reservation.SpotID {
		releaseSpot := bson.M{"$set": bson.M{"status": "available"}}
		_, err := s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, releaseSpot)
		if err != nil {
			return err
		}

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
		log.Println("Error updating reservation:", err)
		return errors.New("failed to update reservation")
	}
	return nil
}

// DeleteReservation releases the parking spot when the reservation is deleted.
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
	updateSpot := bson.M{"$inc": bson.M{"reserved_count": -1}} // Decrement reserved count
	_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, updateSpot)
	if err != nil {
		return errors.New("failed to update parking spot capacity")
	}

	// After decrementing the reserved count, check if the spot is still fully reserved
	spotCount, err := s.ReservationCollection.CountDocuments(ctx, bson.M{"spot_id": reservation.SpotID, "date": reservation.Date})
	if err != nil {
		return errors.New("failed to check reservation count")
	}

	// If the parking spot is no longer fully reserved, mark it as available
	if int(spotCount) < parkingSpot.MaxCapacity {
		log.Printf("SpotID %s is now available\n", reservation.SpotID.Hex())
	}

	// Delete the reservation
	_, err = s.ReservationCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete reservation")
	}

	log.Printf("Reservation for SpotID: %s successfully deleted\n", reservation.SpotID.Hex())
	return nil
}

// AdminDeleteReservation deletes a reservation without user_id restrictions (admin functionality).
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
	updateSpot := bson.M{"$inc": bson.M{"reserved_count": -1}}
	_, err = s.ParkingSpotCollection.UpdateOne(ctx, bson.M{"_id": reservation.SpotID}, updateSpot)
	if err != nil {
		return errors.New("failed to update parking spot capacity")
	}

	// After decrementing the reserved count, check if the spot is still fully reserved
	spotCount, err := s.ReservationCollection.CountDocuments(ctx, bson.M{"spot_id": reservation.SpotID, "date": reservation.Date})
	if err != nil {
		return errors.New("failed to check reservation count")
	}

	if int(spotCount) < parkingSpot.MaxCapacity {
		log.Printf("SpotID %s is now available\n", reservation.SpotID.Hex())
	}

	// Delete the reservation
	_, err = s.ReservationCollection.DeleteOne(ctx, filter)
	if err != nil {
		return errors.New("failed to delete reservation")
	}

	log.Printf("Reservation for SpotID: %s successfully deleted\n", reservation.SpotID.Hex())
	return nil
}
