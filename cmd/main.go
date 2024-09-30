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
	origin = strings.Replace(origin, "http://", "", 1)
	origin = strings.Replace(origin, "https://", "", 1)

	if strings.HasSuffix(origin, ".deiliinvitation.com") || origin == "deiliinvitation.com" {
		return true
	}

	allowedOrigins := []string{
		"https://evelynandbenhard.vercel.app",
		"localhost:3000",
	}

	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}

func main() {
	config.LoadEnv()

	client.Init()
	guest.Init()

	r := mux.NewRouter()

	api.RegisterRoutes(r)

	corsChecker := func(origin string) bool {
		return isAllowedOrigin(origin)
	}

	corsOrigins := handlers.AllowedOriginValidator(corsChecker)
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handlers.CORS(corsOrigins, corsMethods, corsHeaders)(r)))
}
