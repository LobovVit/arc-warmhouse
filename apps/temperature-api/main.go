package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TemperatureResponse struct {
	SensorID string  `json:"sensorId"`
	Location string  `json:"location"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	TS       string  `json:"ts"`
}

func temperatureHandler(w http.ResponseWriter, r *http.Request) {
	location := strings.TrimSpace(r.URL.Query().Get("location"))
	sensorID := strings.TrimSpace(r.URL.Query().Get("sensorId"))
	normalizeLocationAndID(&location, &sensorID)

	writeTemp(w, location, sensorID)
}

func temperatureByIDHandler(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path = "/temperature/1" → берем часть после префикса
	id := strings.TrimPrefix(r.URL.Path, "/temperature/")
	if id == "" || id == "/" {
		http.Error(w, "missing sensorId in path", http.StatusBadRequest)
		return
	}
	// если хвост содержит слэши — отрежем всё после первого
	if i := strings.IndexRune(id, '/'); i >= 0 {
		id = id[:i]
	}
	sensorID := strings.TrimSpace(id)

	// из sensorID определим location по правилам задания
	location := ""
	normalizeLocationAndID(&location, &sensorID)

	writeTemp(w, location, sensorID)
}

func writeTemp(w http.ResponseWriter, location, sensorID string) {
	// рандомная температура (например, 18.0 .. 26.0, одно десятичное)
	val := 18.0 + rand.Float64()*(26.0-18.0)
	val = float64(int(val*10)) / 10.0

	resp := TemperatureResponse{
		SensorID: sensorID,
		Location: location,
		Value:    val,
		Unit:     "C",
		TS:       time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func normalizeLocationAndID(location *string, sensorID *string) {
	// Если location не задан — берем из sensorID
	if *location == "" && *sensorID != "" {
		switch *sensorID {
		case "1":
			*location = "Living Room"
		case "2":
			*location = "Bedroom"
		case "3":
			*location = "Kitchen"
		default:
			*location = "Unknown"
		}
	}

	// Если sensorID не задан — берем из location
	if *sensorID == "" && *location != "" {
		switch *location {
		case "Living Room":
			*sensorID = "1"
		case "Bedroom":
			*sensorID = "2"
		case "Kitchen":
			*sensorID = "3"
		default:
			*sensorID = "0"
		}
	}

	// Дополнительно: если в path пришёл нечисловой id — нормализуем в "0"
	if *sensorID != "" {
		if _, err := strconv.Atoi(*sensorID); err != nil {
			*sensorID = "0"
		}
	}
}

func main() {
	http.HandleFunc("/temperature", temperatureHandler)
	http.HandleFunc("/temperature/", temperatureByIDHandler)
	fmt.Println("🚀 temperature-api running on :8081")
	http.ListenAndServe(":8081", nil)
}
