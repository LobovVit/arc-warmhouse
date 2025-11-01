package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"createdAt"`
}

type Twin struct {
	Version  int64                  `json:"version"`
	Desired  map[string]interface{} `json:"desired"`
	Reported map[string]interface{} `json:"reported"`
	Updated  time.Time              `json:"updatedAt"`
}

type DeviceWithTwin struct {
	Device Device `json:"device"`
	Twin   Twin   `json:"twin"`
}

var (
	mu      sync.RWMutex
	devices = map[string]Device{}
	twins   = map[string]*Twin{}
)

// simpleETag формирует слабый ETag, например: W/"v5"
func simpleETag(version int64) string {
	return `W/"v` + strconv.FormatInt(version, 10) + `"`
}

func jsonNumberFromInt(v int64) string {
	return strings.TrimSpace(
		func() string { b, _ := json.Marshal(v); return string(b) }(),
	)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func listDevices(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Device, 0, len(devices))
	for _, d := range devices {
		out = append(out, d)
	}
	writeJSON(w, http.StatusOK, out)
}

func createDevice(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	mu.Lock()
	defer mu.Unlock()

	id := uuid.NewString()
	d := Device{
		ID:        id,
		Name:      in.Name,
		Type:      in.Type,
		Location:  in.Location,
		CreatedAt: time.Now().UTC(),
	}
	devices[id] = d
	twins[id] = &Twin{
		Version:  1,
		Desired:  map[string]interface{}{},
		Reported: map[string]interface{}{},
		Updated:  time.Now().UTC(),
	}
	writeJSON(w, http.StatusCreated, d)
}

func getTwin(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/devices/")
	id = strings.TrimSuffix(id, "/twin")

	mu.RLock()
	t := twins[id]
	mu.RUnlock()
	if t == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("ETag", simpleETag(t.Version))
	writeJSON(w, http.StatusOK, t)
}

func patchTwin(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/devices/")
	id = strings.TrimSuffix(id, "/twin")

	ifMatch := r.Header.Get("If-Match") // для идемпотентности/конкуренции
	var patch map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	t := twins[id]
	if t == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// проверим версию, если прислали If-Match
	if ifMatch != "" && ifMatch != simpleETag(t.Version) {
		http.Error(w, "version conflict", http.StatusConflict)
		return
	}

	// merge desired (очень простой merge)
	if t.Desired == nil {
		t.Desired = map[string]interface{}{}
	}
	for k, v := range patch {
		t.Desired[k] = v
	}
	t.Version++
	t.Updated = time.Now().UTC()

	w.Header().Set("ETag", simpleETag(t.Version))
	writeJSON(w, http.StatusOK, t)
}

func getDevice(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/devices/")
	mu.RLock()
	defer mu.RUnlock()
	d, ok := devices[id]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	mux := http.NewServeMux()

	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listDevices(w, r)
		case http.MethodPost:
			createDevice(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/devices/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/twin") && r.Method == http.MethodGet:
			getTwin(w, r)
		case strings.HasSuffix(r.URL.Path, "/twin") && r.Method == http.MethodPatch:
			patchTwin(w, r)
		case r.Method == http.MethodGet:
			getDevice(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/heating/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/setpoint") {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/heating/")
		id = strings.TrimSuffix(id, "/setpoint")
		var in struct {
			Value float64 `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Value == 0 {
			http.Error(w, "bad json / value required", http.StatusBadRequest)
			return
		}
		mu.Lock()
		t := twins[id]
		if t == nil {
			mu.Unlock()
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if t.Desired == nil {
			t.Desired = map[string]any{}
		}
		heating, _ := t.Desired["heating"].(map[string]any)
		if heating == nil {
			heating = map[string]any{}
		}
		heating["setpoint"] = in.Value
		t.Desired["heating"] = heating
		t.Version++
		tv := t.Version
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/api/v1/commands/placeholder/status")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":       "accepted",
			"deviceId":     id,
			"desiredPatch": map[string]any{"heating": map[string]any{"setpoint": in.Value}},
			"twinVersion":  tv,
		})
	})

	srv := &http.Server{
		Addr:         ":8082",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Println("🚀 device-service listening on :8082")
	log.Fatal(srv.ListenAndServe())
}
