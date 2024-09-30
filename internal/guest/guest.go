package guest

import (
	"context"
	"deili-backend/config"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Guest struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Name         string             `bson:"name"`
	Message      string             `bson:"message"`
	Confirmation string             `bson:"confirmation"`
	ClientID     primitive.ObjectID `bson:"client_id"`
}

var guestCollection *mongo.Collection
var database *mongo.Database

// Init initializes the MongoDB connection for the guest collection
func Init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Use mongo.Connect() directly
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURI))
	if err != nil {
		log.Fatal("Error connecting to MongoDB:", err)
	}
	database = client.Database(config.DBName)
	guestCollection = database.Collection("guests")
}

// CreateGuest inserts a new guest into the MongoDB guest collection
func CreateGuest(guest Guest) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Validate client_id is not empty
	if guest.ClientID.IsZero() {
		return nil, errors.New("invalid client_id: client_id is zero")
	}

	// Check if the client exists
	clientExists, err := validateClient(guest.ClientID)
	if err != nil {
		return nil, fmt.Errorf("error validating client: %v", err)
	}
	if !clientExists {
		return nil, fmt.Errorf("client with ID %s does not exist", guest.ClientID.Hex())
	}

	result, err := guestCollection.InsertOne(ctx, guest)
	if err != nil {
		return nil, fmt.Errorf("error inserting guest: %v", err)
	}

	return result, nil
}

// validateClient checks if a client with the given ID exists
func validateClient(clientID primitive.ObjectID) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientCollection := database.Collection("clients")
	var client struct{}
	err := clientCollection.FindOne(ctx, bson.M{"_id": clientID}).Decode(&client)
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetGuestsByClient(clientID primitive.ObjectID) ([]Guest, error) {
	var guests []Guest
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := guestCollection.Find(ctx, bson.M{"client_id": clientID})
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &guests); err != nil {
		return nil, err
	}

	return guests, nil
}

// GetGuestByID retrieves a guest by its ObjectID
func GetGuestByID(id primitive.ObjectID) (*Guest, error) {
	var guest Guest
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := guestCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&guest)
	return &guest, err
}

// UpdateGuest updates an existing guest's information
func UpdateGuest(id primitive.ObjectID, updatedData Guest) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate client_id is not empty
	if updatedData.ClientID.IsZero() {
		return nil, errors.New("invalid client_id: client_id is zero")
	}

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"name":         updatedData.Name,
			"message":      updatedData.Message,
			"confirmation": updatedData.Confirmation,
			"client_id":    updatedData.ClientID,
		},
	}

	return guestCollection.UpdateOne(ctx, filter, update)
}

// DeleteGuest deletes a guest from the collection based on its ObjectID
func DeleteGuest(id primitive.ObjectID) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return guestCollection.DeleteOne(ctx, bson.M{"_id": id})
}
