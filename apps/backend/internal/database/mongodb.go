package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoClient represents the MongoDB client connection
type MongoClient struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient(uri, dbName string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the MongoDB server to verify connection
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")

	// Get a handle to the specified database
	db := client.Database(dbName)

	return &MongoClient{
		Client: client,
		DB:     db,
	}, nil
}

// Close disconnects from MongoDB
func (m *MongoClient) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

// GetCollection returns a handle to the specified collection
func (m *MongoClient) GetCollection(name string) *mongo.Collection {
	return m.DB.Collection(name)
} 