package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// Database interface defines the operations needed by handlers
type Database interface {
	Collection(name string) *mongo.Collection
	Ping(ctx context.Context) error
	GetMongoDB() *mongo.Database
}

// MongoDB is a wrapper around *mongo.Database that implements the Database interface
type MongoDB struct {
	db *mongo.Database
}

// NewMongoDB creates a new MongoDB instance that implements the Database interface
func NewMongoDB(db *mongo.Database) Database {
	return &MongoDB{db: db}
}

// Collection returns a handle to a collection
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.db.Collection(name)
}

// Ping checks the database connection
func (m *MongoDB) Ping(ctx context.Context) error {
	return m.db.Client().Ping(ctx, nil)
}

// GetMongoDB returns the underlying MongoDB database instance
func (m *MongoDB) GetMongoDB() *mongo.Database {
	return m.db
}
