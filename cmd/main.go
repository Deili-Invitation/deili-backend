// package main

// import (
// 	"deili-backend/api"
// 	"deili-backend/config"
// 	"deili-backend/internal/client"
// 	"deili-backend/internal/guest"
// 	"log"
// 	"net/http"
// 	"os"
// 	"strings"

// 	"github.com/gorilla/handlers"
// 	"github.com/gorilla/mux"
// )

// func isAllowedOrigin(origin string) bool {
// 	allowedOrigins := []string{
// 		"https://evelynandbenhard.vercel.app",
// 		"https://evelynandbenhard.deiliinvitation.com",
// 		"http://localhost:3000",
// 		"https://localhost:3000",
// 	}
// 	for _, allowedOrigin := range allowedOrigins {
// 		if origin == allowedOrigin {
// 			return true
// 		}
// 	}
// 	return strings.HasSuffix(origin, ".deiliinvitation.com") || origin == "https://deiliinvitation.com"
// }

// func main() {
// 	config.LoadEnv()
// 	client.Init()
// 	guest.Init()

// 	r := mux.NewRouter()
// 	api.RegisterRoutes(r)

// 	corsMiddleware := handlers.CORS(
// 		handlers.AllowedOrigins([]string{"http://localhost:3000", "https://deiliinvitation.com", ".deiliinvitation.com", "https://evelynandbenhard.deiliinvitation.com"}),
// 		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
// 		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
// 		handlers.AllowCredentials(),
// 	)

// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080"
// 	}

//		log.Printf("Server is running on port %s", port)
//		log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(r)))
//	}
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"deili-backend/api"
	"deili-backend/config"
	"deili-backend/internal/client"
	"deili-backend/internal/guest"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// isAllowedOrigin checks if the request origin is allowed for CORS.
func isAllowedOrigin(origin string) bool {
	// Allow main domain and any subdomain under `deiliinvitation.com`
	if origin == "https://deiliinvitation.com" || strings.HasSuffix(origin, ".deiliinvitation.com") {
		return true
	}
	// Allow localhost for local development
	return origin == "http://localhost:3000" || origin == "https://localhost:3000"
}

// connectMongoDB establishes a connection to MongoDB using the URI from Render's environment.
func connectMongoDB() *mongo.Client {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the MongoDB server to verify the connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Could not ping MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB!")
	return client
}

func main() {
	// Load environment configuration, initialize client and guest modules
	config.LoadEnv()
	client.Init()
	guest.Init()

	// Initialize MongoDB connection
	dbClient := connectMongoDB()
	defer func() {
		if err := dbClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Set up the router and register API routes
	r := mux.NewRouter()
	api.RegisterRoutes(r)

	// Set up CORS middleware with dynamic origin validation for subdomains and main domain
	corsMiddleware := handlers.CORS(
		handlers.AllowedOriginValidator(isAllowedOrigin),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	// Retrieve PORT from environment, default to 8080 if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server with CORS middleware applied to the router
	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(r)))
}
