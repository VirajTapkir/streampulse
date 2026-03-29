package scoring

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/VirajTapkir/streampulse/db"
)

// MomentumScore holds the computed score and its components
// so the dashboard can show both the final number and the breakdown
type MomentumScore struct {
	Score       float64 `json:"score"`        // the final weighted score
	SubRate     float64 `json:"sub_rate"`     // subs per minute
	BitsPerMin  float64 `json:"bits_per_min"` // bits events per minute
	ChatDensity float64 `json:"chat_density"` // donations per minute
	ComputedAt  string  `json:"computed_at"`  // when this score was calculated
}

// StartMomentumTicker runs a goroutine that recomputes the score
// every 5 seconds and broadcasts it to all connected clients
// hub is passed as an interface so we avoid an import cycle
func StartMomentumTicker(broadcast chan<- []byte) {
	go func() {
		// time.Tick returns a channel that fires every 5 seconds
		// each tick is a signal to recompute the score
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			score, err := Compute()
			if err != nil {
				fmt.Println("momentum compute error:", err)
				continue
			}

			// marshal the score to JSON and send it to the broadcast channel
			payload, err := json.Marshal(score)
			if err != nil {
				continue
			}

			// wrap it so the frontend knows this is a momentum message
			// not a regular event like sub/bits/donation
			wrapped, _ := json.Marshal(map[string]interface{}{
				"type":    "momentum",
				"payload": score,
			})

			// store the latest score in Redis so the REST API can serve it
			db.RDB.Set(context.Background(), "momentum:score", payload, 10*time.Second)

			// send to broadcast channel — hub will fan it out to all clients
			broadcast <- wrapped
		}
	}()
}

// Compute reads the Redis counters and calculates the momentum score
// it divides by 5 because our window is 5 seconds = 1/12th of a minute
// so we multiply by 12 to get a per-minute rate
func Compute() (MomentumScore, error) {
	ctx := context.Background()

	// read the three counters from Redis
	subs, err := db.RDB.Get(ctx, "counter:sub").Float64()
	if err != nil {
		subs = 0 // key doesn't exist yet — treat as zero
	}

	bits, err := db.RDB.Get(ctx, "counter:bits").Float64()
	if err != nil {
		bits = 0
	}

	donations, err := db.RDB.Get(ctx, "counter:donation").Float64()
	if err != nil {
		donations = 0
	}

	// convert raw counts to per-minute rates
	// we check every 5 seconds so multiply by 12 to get per-minute
	subRate := subs * 12
	bitsPerMin := bits * 12
	chatDensity := donations * 12

	// apply the weighted formula
	score := (subRate * 0.5) + (bitsPerMin * 0.3) + (chatDensity * 0.2)

	// reset counters so the next 5-second window starts fresh
	// without this, counts would grow forever and rates would be wrong
	db.RDB.Set(ctx, "counter:sub", 0, 0)
	db.RDB.Set(ctx, "counter:bits", 0, 0)
	db.RDB.Set(ctx, "counter:donation", 0, 0)

	return MomentumScore{
		Score:       score,
		SubRate:     subRate,
		BitsPerMin:  bitsPerMin,
		ChatDensity: chatDensity,
		ComputedAt:  time.Now().Format(time.RFC3339),
	}, nil
}		