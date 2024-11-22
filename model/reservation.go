package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Reservation struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SpotID      primitive.ObjectID `json:"spot_id,omitempty" bson:"spot_id,omitempty"`             // References ParkingSpot
	FoodTruckID primitive.ObjectID `json:"food_truck_id,omitempty" bson:"food_truck_id,omitempty"` // References FoodTruck
	UserID      primitive.ObjectID `json:"user_id,omitempty" bson:"user_id,omitempty"`             // References User
	Date        time.Time          `json:"date,omitempty" bson:"date,omitempty"`                   // Reservation date
	CreatedAt   time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`       // Reservation creation date
}
