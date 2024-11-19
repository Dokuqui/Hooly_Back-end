package db

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var Database *mongo.Database

func Connect() {
	// Create a context with 10-seconds timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DB_NAME") // Set in your .env file

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	// Test connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB: ", err)
	}

	log.Println("Connected to MongoDB")

	// Store the client for future use
	Client = client
	Database = client.Database(dbName) // Set the DB name here
}

// GetCollection returns a collection for the given name from the connected DB
func GetCollection(name string) *mongo.Collection {
	return Database.Collection(name)
}
