package config

import (
	"log"
	"os"
)

// MongoURI and DBName are global variables that hold the database URI and database name.
var MongoURI string
var DBName string

// LoadEnv retrieves environment variables for MongoDB configuration.
func LoadEnv() {
	// Get MongoDB connection URI from the environment
	MongoURI = os.Getenv("MONGO_URI")
	if MongoURI == "" {
		log.Fatal("MONGO_URI environment variable is not set")
	}

	// Get the database name from the environment
	DBName = os.Getenv("DB_NAME")
	if DBName == "" {
		log.Fatal("DB_NAME environment variable is not set")
	}
}
