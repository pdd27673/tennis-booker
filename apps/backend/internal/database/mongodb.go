package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient wraps the MongoDB client with additional functionality
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoClient creates a new MongoDB client connection
func NewMongoClient(uri, databaseName string) (*MongoClient, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Successfully connected to MongoDB

	// Get a handle to the specified database
	database := client.Database(databaseName)

	return &MongoClient{
		client:   client,
		database: database,
	}, nil
}

// GetClient returns the underlying MongoDB client
func (m *MongoClient) GetClient() *mongo.Client {
	return m.client
}

// GetDatabase returns the MongoDB database
func (m *MongoClient) GetDatabase() *mongo.Database {
	return m.database
}

// Close closes the MongoDB connection
func (m *MongoClient) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// GetCollection returns a handle to the specified collection
func (m *MongoClient) GetCollection(name string) *mongo.Collection {
	return m.database.Collection(name)
}
