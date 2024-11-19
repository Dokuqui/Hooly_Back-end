package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Foodtruck struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Name   string             `json:"name" bson:"name"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`
}
