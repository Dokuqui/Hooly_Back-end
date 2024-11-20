package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type ParkingSpot struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Day         string             `bson:"day_of_week" json:"day_of_week"`
	MaxCapacity int                `bson:"max_capacity" json:"max_capacity"`
	Reserved    bool               `bson:"reserved" json:"reserved"`
}
