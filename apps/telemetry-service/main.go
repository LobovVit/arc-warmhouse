package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Telemetry struct {
	DeviceID  string    `json:"deviceId"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit,omitempty"`
	TS        time.Time `json:"ts"`
	MessageID string    `json:"messageId,omitempty"`
}

var (
	mu   sync.RWMutex
	data = map[string][]Telemetry{} // map[deviceId][]Telemetry (in-memory)
)

func postIngest(w http.ResponseWriter, r *http.Request) {
	var t Telemetry
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if t.DeviceID == "" || t.Metric == "" {
		http.Error(w, "deviceId and metric required", http.StatusBadRequest)
		return
	}
	if t.TS.IsZero() {
		t.TS = time.Now().UTC()
	}
	if t.Value == 0 { // для демо — рандом если не передали
		t.Value = 18 + rand.Float64()*(27-18)
	}

	mu.Lock()
	data[t.DeviceID] = append(data[t.DeviceID], t)
	mu.Unlock()

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "accepted",
		"count":  len(data[t.DeviceID]),
	})
}

func getTelemetry(w http.ResponseWriter, r *http.Request) {
	deviceID := strings.TrimPrefix(r.URL.Path, "/telemetry/")
	metric := r.URL.Query().Get("metric")
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}

	mu.RLock()
	defer mu.RUnlock()
	arr := data[deviceID]
	out := []Telemetry{}
	for i := len(arr) - 1; i >= 0 && len(out) < limit; i-- {
		if metric == "" || arr[i].Metric == metric {
			out = append(out, arr[i])
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	mux := http.NewServeMux()

	mux.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		postIngest(w, r)
	})
	mux.HandleFunc("/telemetry/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		getTelemetry(w, r)
	})

	addr := ":8083"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Println("📡 telemetry-service listening on", addr)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
