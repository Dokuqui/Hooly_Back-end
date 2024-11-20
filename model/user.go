package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Firstname string             `json:"firstname" bson:"firstname"`
	Lastname  string             `json:"lastname" bson:"lastname"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-" validate:"required,min=6"` // Store hashed password
	Role      string             `bson:"role" json:"role"`                            // Role can be "admin" or "user"
}
