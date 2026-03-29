package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/VirajTapkir/streampulse/db"
)

// respondJSON is a helper that writes any Go value as a JSON response
// we use this in every handler so we don't repeat ourselves
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// GetStreamers handles GET /api/streamers
// It reads all rows from the streamers table and returns them as JSON
func GetStreamers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, username, display_name, created_at FROM streamers")
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to query streamers",
		})
		return
	}
	defer rows.Close() // always close rows when done to free the connection

	// build a slice to hold all the results
	type Streamer struct {
		ID          int    `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}

	streamers := []Streamer{}


	for rows.Next() {
		var s Streamer
		var createdAt interface{} // we read this but don't need it in the response
		if err := rows.Scan(&s.ID, &s.Username, &s.DisplayName, &createdAt); err != nil {
			continue
		}
		streamers = append(streamers, s)
	}

	respondJSON(w, http.StatusOK, streamers)
}

// GetEarnings handles GET /api/earnings
// It reads all rows from the earnings table and returns them as JSON
func GetEarnings(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(
		"SELECT id, streamer_id, event_type, amount, occurred_at FROM earnings ORDER BY occurred_at DESC LIMIT 50",
	)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to query earnings",
		})
		return
	}
	defer rows.Close()

	type Earning struct {
		ID         int     `json:"id"`
		StreamerID int     `json:"streamer_id"`
		EventType  string  `json:"event_type"`
		Amount     float64 `json:"amount"`
	}

	earnings := []Earning{}


	for rows.Next() {
		var e Earning
		var occurredAt interface{}
		if err := rows.Scan(&e.ID, &e.StreamerID, &e.EventType, &e.Amount, &occurredAt); err != nil {
			continue
		}
		earnings = append(earnings, e)
	}

	respondJSON(w, http.StatusOK, earnings)
}
// GetMomentum handles GET /api/momentum
// returns the latest momentum score stored in Redis
func GetMomentum(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	val, err := db.RDB.Get(ctx, "momentum:score").Result()
	if err != nil {
		// score hasn't been computed yet — return a zero score
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"score":       0,
			"sub_rate":    0,
			"bits_per_min": 0,
			"chat_density": 0,
			"computed_at": "not yet computed",
		})
		return
	}

	// the score is already JSON — write it directly
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(val))
}