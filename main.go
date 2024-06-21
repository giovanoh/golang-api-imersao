package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type MessageResponse struct {
	Message string `json:"message"`
}

type Event struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Organization string  `json:"organization"`
	Date         string  `json:"date"`
	Price        float64 `json:"price"`
	Rating       string  `json:"rating"`
	ImageURL     string  `json:"image_url"`
	CreatedAt    string  `json:"created_at"`
	Location     string  `json:"location"`
}

type Spot struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	EventID int    `json:"event_id"`
}

type EventStore struct {
	Events []Event `json:"events"`
	Spots  []Spot  `json:"spots"`
}

var store EventStore

func main() {
	var err error
	store, err = LoadJSON[EventStore]("./data.json")
	if err != nil {
		panic(err)
	}

	server := http.NewServeMux()
	server.HandleFunc("/events", ListEvents)
	server.HandleFunc("/events/{eventID}", GetEvent)
	server.HandleFunc("/events/{eventID}/spots", ListSpots)
	server.HandleFunc("POST /events/{eventID}/reserve", ReserveSpot)

	http.ListenAndServe(":8080", server)
}

func LoadJSON[T any](filename string) (T, error) {
	var data T
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return data, err
	}
	return data, json.Unmarshal(fileData, &data)
}

func writeResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(MessageResponse{Message: message})
}

func ListEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(store.Events)
}

func GetEvent(w http.ResponseWriter, r *http.Request) {
	eventId := r.PathValue("eventID")
	if eventId == "" {
		writeResponse(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	intEventId, err := strconv.Atoi(eventId)
	if err != nil {
		writeResponse(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	var event Event
	for _, e := range store.Events {
		if e.ID == intEventId {
			event = e
			break
		}
	}
	if event == (Event{}) {
		writeResponse(w, "Event not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func ListSpots(w http.ResponseWriter, r *http.Request) {
	eventId := r.PathValue("eventID")
	if eventId == "" {
		writeResponse(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	intEventId, err := strconv.Atoi(eventId)
	if err != nil {
		writeResponse(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	var event Event
	for _, e := range store.Events {
		if e.ID == intEventId {
			event = e
			break
		}
	}
	if event == (Event{}) {
		writeResponse(w, "Event not found", http.StatusNotFound)
		return
	}

	var spots []Spot
	for _, spot := range store.Spots {
		if spot.EventID == event.ID {
			spots = append(spots, spot)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spots)
}

func ReserveSpot(w http.ResponseWriter, r *http.Request) {
	eventId := r.PathValue("eventID")
	if eventId == "" {
		writeResponse(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	intEventId, err := strconv.Atoi(eventId)
	if err != nil {
		writeResponse(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Parse body
	var spotParams []string
	err = json.NewDecoder(r.Body).Decode(&spotParams)
	if err != nil {
		writeResponse(w, "Invalid request body. Expected an array of strings with the spot names", http.StatusBadRequest)
		return
	}

	// Checar se o evento existe
	var event Event
	for _, e := range store.Events {
		if e.ID == intEventId {
			event = e
			break
		}
	}
	if event == (Event{}) {
		writeResponse(w, "Event not found", http.StatusNotFound)
		return
	}

	// Checar se spot existe
	var notFoundSpots []string
	for _, spotName := range spotParams {
		found := false
		for _, spot := range store.Spots {
			if spot.EventID == event.ID && spot.Name == spotName {
				found = true
				break
			}
		}
		if !found {
			notFoundSpots = append(notFoundSpots, spotName)
		}
	}
	if len(notFoundSpots) > 0 {
		writeResponse(w, fmt.Sprint("Spot ", strings.Join(notFoundSpots, ", "), " not found"), http.StatusNotFound)
		return
	}

	// Checar se spot já está reservado
	var reservedSpots []string
	for _, spotName := range spotParams {
		for _, spot := range store.Spots {
			if spot.EventID != event.ID {
				continue
			}
			if spot.Name == spotName && spot.Status == "reserved" {
				reservedSpots = append(reservedSpots, spotName)
			}
		}
	}
	if len(reservedSpots) > 0 {
		writeResponse(w, fmt.Sprint("Spot ", strings.Join(reservedSpots, ", "), " already reserved"), http.StatusBadRequest)
		return
	}

	// Reserver spots
	for _, spotName := range spotParams {
		for i, spot := range store.Spots {
			if spot.EventID == event.ID && spot.Name == spotName {
				store.Spots[i].Status = "reserved" // Atualizando pelo índice porque o range retorna uma cópia do dado
			}
		}
	}

	writeResponse(w, "Spots reserved successfully", http.StatusCreated)
}
