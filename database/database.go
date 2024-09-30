package database

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToMongoDB() (*mongo.Client, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI environment variable is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	clientOptions := options.Client().ApplyURI(uri).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(15 * time.Second).
		SetSocketTimeout(30 * time.Second).
		SetTLSConfig(tlsConfig)

	var client *mongo.Client
	var err error

	for retries := 0; retries < 5; retries++ {
		client, err = mongo.Connect(ctx, clientOptions)
		if err == nil {
			err = client.Ping(ctx, nil)
			if err == nil {
				log.Printf("Successfully connected to MongoDB on attempt %d", retries+1)
				return client, nil
			}
		}
		log.Printf("Failed to connect to MongoDB on attempt %d: %v", retries+1, err)
		time.Sleep(time.Duration(retries+1) * 2 * time.Second)
	}

	return nil, err
}
