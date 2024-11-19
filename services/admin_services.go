package services

import (
	"context"
	"gitlab.com/hooly2/back/db"
	"gitlab.com/hooly2/back/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type LogService struct {
	LogCollection *mongo.Collection
}

type MonitoringService struct {
	ReservationCollection *mongo.Collection
	LogCollection         *mongo.Collection
}

func NewLogService() *LogService {
	return &LogService{
		LogCollection: db.GetCollection("log"),
	}
}

func NewMonitoringService() *MonitoringService {
	return &MonitoringService{
		ReservationCollection: db.GetCollection("reservation"),
		LogCollection:         db.GetCollection("log"),
	}
}

// CreateLog stores a new log entry in the database
func (ls *LogService) CreateLog(level, action, userID, message string) error {
	logEntry := model.Log{
		Timestamp: time.Now(),
		Level:     level,
		Action:    action,
		UserID:    userID,
		Message:   message,
	}

	_, err := ls.LogCollection.InsertOne(context.TODO(), logEntry)
	return err
}

// FetchLogs retrieves logs filtered by action, level, or date
func (ls *LogService) FetchLogs(filter map[string]interface{}) ([]model.Log, error) {
	// Query the database using the provided filters
	cursor, err := ls.LogCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	// Decode logs into a slice
	var logs []model.Log
	if err := cursor.All(context.TODO(), &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// FetchMonitoringData aggregates monitoring data for the admin dashboard
func (ms *MonitoringService) FetchMonitoringData() (model.Monitoring, error) {
	var monitoringData model.Monitoring

	// Count total reservations
	totalReservations, err := ms.ReservationCollection.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		return monitoringData, err
	}
	monitoringData.TotalReservations = int(totalReservations)

	// Count available spots (total spots - total reservations)
	totalSpots := 7*6 + 6 // 7 days of the week, 6 on Friday
	monitoringData.AvailableSpots = totalSpots - int(totalReservations)

	// Count logged errors
	errorCount, err := ms.LogCollection.CountDocuments(context.TODO(), bson.M{"level": "ERROR"})
	if err != nil {
		return monitoringData, err
	}
	monitoringData.ErrorsLogged = int(errorCount)

	return monitoringData, nil
}
