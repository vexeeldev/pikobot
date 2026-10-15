package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

var (
	instance *MongoDB
	once     sync.Once
	initErr  error
)

func Connect(uri, dbName string) (*MongoDB, error) {
	if uri == "" {
		return nil, fmt.Errorf("MONGODB_URI is empty")
	}

	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().
			ApplyURI(uri).
			SetMaxPoolSize(10).
			SetMinPoolSize(1).
			SetMaxConnIdleTime(60 * time.Second)

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			initErr = fmt.Errorf("failed to create MongoDB client: %w", err)
			return
		}

		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			initErr = fmt.Errorf("failed to ping MongoDB: %w", err)
			return
		}

		log.Printf("🍃 Connected successfully to MongoDB database: %s", dbName)

		instance = &MongoDB{
			Client:   client,
			Database: client.Database(dbName),
		}
	})

	return instance, initErr
}

func GetInstance() *MongoDB {
	return instance
}

func (m *MongoDB) Ping(ctx context.Context) error {
	if m == nil || m.Client == nil {
		return fmt.Errorf("mongodb client is nil")
	}
	return m.Client.Ping(ctx, readpref.Primary())
}

func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m == nil || m.Client == nil {
		return nil
	}
	return m.Client.Disconnect(ctx)
}
