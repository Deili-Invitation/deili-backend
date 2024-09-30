package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var MongoURI string
var DBName string

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get MongoDB connection URI from the environment
	MongoURI = os.Getenv("MONGO_URI")
	if MongoURI == "" {
		log.Fatal("MONGO_URI not set in .env file")
	}

	DBName = os.Getenv("DB_NAME")
	if DBName == "" {
		log.Fatal("DB_NAME not set in .env file")
	}
}
