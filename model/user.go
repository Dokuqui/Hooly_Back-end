package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Firstname string             `json:"firstname" bson:"firstname"`
	Lastname  string             `json:"lastname" bson:"lastname"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"-" validate:"required,min=6"` // Store hashed password
	Role      Role               `bson:"role" json:"role"`                            // Role can be "admin" or "user"
}
