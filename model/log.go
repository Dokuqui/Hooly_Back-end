package model

import "time"

type Log struct {
	ID        string    `bson:"_id,omitempty"` // MongoDB ObjectID
	Timestamp time.Time `bson:"timestamp"`     // Time of the logged action
	Level     string    `bson:"level"`         // INFO, ERROR, WARN
	Action    string    `bson:"action"`        // Reservation, Deletion, etc.
	UserID    string    `bson:"userId"`        // User who performed the action
	Message   string    `bson:"message"`       // Details of the log
}
