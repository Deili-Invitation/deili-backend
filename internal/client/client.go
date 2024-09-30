package client

import (
	"context"
	"deili-backend/config"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client struct represents the client data structure
type Client struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name            string             `bson:"name" json:"name"`
	Contact         string             `bson:"contact" json:"contact"`
	InvitationTypes string             `bson:"invitation_types" json:"invitation_types"`
}

var clientCollection *mongo.Collection

// Init initializes the MongoDB connection for the client collection
func Init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use mongo.Connect() directly instead of deprecated mongo.NewClient()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURI))
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}

	clientCollection = client.Database(config.DBName).Collection("clients")
}

// CreateClient inserts a new client into the MongoDB client collection
func CreateClient(client Client) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return clientCollection.InsertOne(ctx, client)
}

// GetClients retrieves all clients from the MongoDB client collection
func GetClients() ([]Client, error) {
	var clients []Client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := clientCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &clients); err != nil {
		return nil, err
	}

	return clients, nil
}

// GetClientByID retrieves a client by its ObjectID
func GetClientByID(id primitive.ObjectID) (*Client, error) {
	var client Client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := clientCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&client)
	return &client, err
}

// UpdateClient updates only the fields provided in the request body for a client
func UpdateClient(id primitive.ObjectID, updatedData map[string]interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": updatedData, // Dynamically set only the fields that are provided
	}
	return clientCollection.UpdateOne(ctx, filter, update)
}

// DeleteClient deletes a client from the collection based on its ObjectID
func DeleteClient(id primitive.ObjectID) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return clientCollection.DeleteOne(ctx, bson.M{"_id": id})
}
