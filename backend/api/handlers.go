package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/VirajTapkir/streampulse/db"
)

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func GetStreamers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, username, display_name, created_at FROM streamers")
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to query streamers",
		})
		return
	}
	defer rows.Close() 

	type Streamer struct {
		ID          int    `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	}

	streamers := []Streamer{}


	for rows.Next() {
		var s Streamer
		var createdAt interface{} 
		if err := rows.Scan(&s.ID, &s.Username, &s.DisplayName, &createdAt); err != nil {
			continue
		}
		streamers = append(streamers, s)
	}

	respondJSON(w, http.StatusOK, streamers)
}

func GetEarnings(w http.ResponseWriter, r *http.Request) {
	streamerID := "1"
	if sid := r.URL.Query().Get("streamer_id"); sid != "" {
		streamerID = sid
	}

	rows, err := db.DB.Query(
		"SELECT id, streamer_id, event_type, amount, occurred_at FROM earnings WHERE streamer_id = $1 ORDER BY occurred_at DESC LIMIT 50",
		streamerID,
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

func GetMomentum(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	streamerID := "1"
	if sid := r.URL.Query().Get("streamer_id"); sid != "" {
		streamerID = sid
	}

	val, err := db.RDB.Get(ctx, fmt.Sprintf("streamer:%s:momentum:score", streamerID)).Result()

	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"score":       0,
			"sub_rate":    0,
			"bits_per_min": 0,
			"chat_density": 0,
			"computed_at": "not yet computed",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(val))
}
func GetAnalytics(w http.ResponseWriter, r *http.Request) {
	streamerID := "1"
	if sid := r.URL.Query().Get("streamer_id"); sid != "" {
		streamerID = sid
	}

	days := "7"
	if d := r.URL.Query().Get("days"); d != "" {
		days = d
	}

	rows, err := db.DB.Query(`
		SELECT
			DATE_TRUNC('day', occurred_at) AS day,
			event_type,
			SUM(amount)                    AS total
		FROM earnings
		WHERE
			streamer_id = $1
			AND occurred_at >= NOW() - ($2 || ' days')::INTERVAL
		GROUP BY day, event_type
		ORDER BY day ASC
	`, streamerID, days)

	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to query analytics",
		})
		return
	}
	defer rows.Close()

	type DailyTotal struct {
		Day       string  `json:"day"`
		EventType string  `json:"event_type"`
		Total     float64 `json:"total"`
	}

	results := []DailyTotal{}

	for rows.Next() {
		var d DailyTotal
		var day time.Time
		if err := rows.Scan(&day, &d.EventType, &d.Total); err != nil {
			continue
		}
		d.Day = day.Format("Jan 2")
		results = append(results, d)
	}

	respondJSON(w, http.StatusOK, results)
}