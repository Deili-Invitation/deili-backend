package api

import (
	"bytes"
	"deili-backend/internal/client"
	"deili-backend/internal/guest"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRoutes(r *mux.Router) {
	// Client routes
	r.HandleFunc("/clients", CreateClient).Methods("POST")
	r.HandleFunc("/clients", GetClients).Methods("GET")
	r.HandleFunc("/clients/{id}", GetClientByID).Methods("GET")
	r.HandleFunc("/clients/{id}", UpdateClient).Methods("PUT")
	r.HandleFunc("/clients/{id}", DeleteClient).Methods("DELETE")

	// Guest routes
	r.HandleFunc("/guests", CreateGuest).Methods("POST")
	r.HandleFunc("/guests/{id}", GetGuestByID).Methods("GET")
	r.HandleFunc("/guests", GetGuestsByClient).Methods("GET")
	r.HandleFunc("/guests/{id}", UpdateGuest).Methods("PUT")
	r.HandleFunc("/guests/{id}", DeleteGuest).Methods("DELETE")
}

func CreateClient(w http.ResponseWriter, r *http.Request) {
	var newClient client.Client
	if err := json.NewDecoder(r.Body).Decode(&newClient); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	// Log the received client data
	log.Printf("Received client data: %+v", newClient)

	// Validate InvitationTypes
	if newClient.InvitationTypes == "" {
		log.Printf("InvitationTypes is empty, request will be rejected")
		http.Error(w, "InvitationTypes cannot be empty", http.StatusBadRequest)
		return
	}

	// Insert the new client into the database
	result, err := client.CreateClient(newClient)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create client: %v", err), http.StatusInternalServerError)
		return
	}

	// Log the created client result
	log.Printf("Created client result: %+v", result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetClients retrieves all clients
func GetClients(w http.ResponseWriter, r *http.Request) {
	clients, err := client.GetClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(clients)
}

// GetClientByID retrieves a client by its ObjectID
func GetClientByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clientID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	clientData, err := client.GetClientByID(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(clientData)
}

// UpdateClient handles updating a client by ID, only the fields provided will be updated
func UpdateClient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clientID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse the update data from the request body as a dynamic map
	var updatedData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updatedData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the client with only the fields provided in the request body
	result, err := client.UpdateClient(clientID, updatedData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// DeleteClient handles deleting a client by ID
func DeleteClient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clientID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := client.DeleteClient(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

// Guest Handlers

func CreateGuest(w http.ResponseWriter, r *http.Request) {
	// Log the raw request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	log.Printf("Raw request body: %s", string(bodyBytes))

	// Restore the request body for further processing
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var newGuest guest.Guest
	if err := json.NewDecoder(r.Body).Decode(&newGuest); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	// Extract client_id from the request body
	var requestBody map[string]interface{}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the body for second read
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	clientIDStr, ok := requestBody["client_id"].(string)
	if !ok {
		log.Printf("client_id is missing or not a string")
		http.Error(w, "client_id must be a valid string", http.StatusBadRequest)
		return
	}

	// Convert client_id from string to ObjectID
	clientID, err := primitive.ObjectIDFromHex(clientIDStr)
	if err != nil {
		log.Printf("Error converting client_id to ObjectID: %v", err)
		http.Error(w, fmt.Sprintf("Invalid client_id: %v", err), http.StatusBadRequest)
		return
	}

	newGuest.ClientID = clientID

	// Insert the new guest into the database
	result, err := guest.CreateGuest(newGuest)
	if err != nil {
		log.Printf("Error creating guest: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create guest: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func GetGuestsByClient(w http.ResponseWriter, r *http.Request) {
	// Fetch clientID from the URL query parameters
	clientIDHex := r.URL.Query().Get("client_id")
	if clientIDHex == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	// Convert the client_id from string to MongoDB ObjectID
	clientID, err := primitive.ObjectIDFromHex(clientIDHex)
	if err != nil {
		log.Printf("Invalid client_id format: %v", err)
		http.Error(w, "Invalid client_id format", http.StatusBadRequest)
		return
	}

	// Fetch guests associated with the given clientID
	guests, err := guest.GetGuestsByClient(clientID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("No guests found for client ID: %s", clientIDHex)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]guest.Guest{}) // Return an empty array instead of error
			return
		}
		log.Printf("Error fetching guests: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the guests found
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guests)
}

func GetGuestByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	guestID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	guestData, err := guest.GetGuestByID(guestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(guestData)
}

func UpdateGuest(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	guestID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		log.Printf("Invalid guest ID: %v", err)
		http.Error(w, "Invalid guest ID", http.StatusBadRequest)
		return
	}

	var updatedGuest guest.Guest
	if err := json.NewDecoder(r.Body).Decode(&updatedGuest); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the existing guest to get the current ClientID if not provided in the request
	existingGuest, err := guest.GetGuestByID(guestID)
	if err != nil {
		log.Printf("Error fetching existing guest: %v", err)
		http.Error(w, "Failed to fetch existing guest", http.StatusInternalServerError)
		return
	}

	// If the ClientID is not provided, use the existing ClientID
	if updatedGuest.ClientID.IsZero() {
		updatedGuest.ClientID = existingGuest.ClientID
	}

	// Validate that the ClientID is not zero
	if updatedGuest.ClientID.IsZero() {
		log.Printf("Invalid client_id: client_id is zero")
		http.Error(w, "Invalid client_id: client_id cannot be zero", http.StatusBadRequest)
		return
	}

	// Update the guest with the new or existing data
	result, err := guest.UpdateGuest(guestID, updatedGuest)
	if err != nil {
		log.Printf("Error updating guest: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func DeleteGuest(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	guestID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := guest.DeleteGuest(guestID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}
