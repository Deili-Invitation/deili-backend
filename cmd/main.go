package main

import (
	"deili-backend/api"
	"deili-backend/config"
	"deili-backend/internal/client"
	"deili-backend/internal/guest"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func isAllowedOrigin(origin string) bool {
	allowedOrigins := []string{
		"https://evelynandbenhard.vercel.app",
		"https://evelynandbenhard.deiliinvitation.com",
		"http://localhost:3000",
		"https://localhost:3000",
	}
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}
	return strings.HasSuffix(origin, ".deiliinvitation.com") || origin == "https://deiliinvitation.com"
}

func main() {
	config.LoadEnv()
	client.Init()
	guest.Init()

	r := mux.NewRouter()
	api.RegisterRoutes(r)

	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"https://evelynandbenhard.vercel.app", "http://localhost:3000", "https://deiliinvitation.com", ".deiliinvitation.com", "https://evelynandbenhard.deiliinvitation.com"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(r)))
}
