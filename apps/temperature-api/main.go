package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type TemperatureResponse struct {
	Location  string  `json:"location"`
	SensorID  string  `json:"sensorId"`
	Value     float64 `json:"temperature"`
	Timestamp string  `json:"timestamp"`
}

func temperatureHandler(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	sensorID := r.URL.Query().Get("sensorId")

	// Default mapping from sensorID → location
	if location == "" {
		switch sensorID {
		case "1":
			location = "Living Room"
		case "2":
			location = "Bedroom"
		case "3":
			location = "Kitchen"
		default:
			location = "Unknown"
		}
	}

	// Default mapping from location → sensorID
	if sensorID == "" {
		switch location {
		case "Living Room":
			sensorID = "1"
		case "Bedroom":
			sensorID = "2"
		case "Kitchen":
			sensorID = "3"
		default:
			sensorID = "0"
		}
	}

	// Generate random temperature between 18–27 °C
	rand.New(rand.NewSource(time.Now().UnixNano()))
	value := 18 + rand.Float64()*(27-18)

	resp := TemperatureResponse{
		Location:  location,
		SensorID:  sensorID,
		Value:     value,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/temperature", temperatureHandler)
	fmt.Println("🚀 temperature-api running on :8081")
	http.ListenAndServe(":8081", nil)
}
