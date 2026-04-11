package scoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/VirajTapkir/streampulse/db"
)


type MomentumScore struct {
	Score       float64 `json:"score"`
	SubRate     float64 `json:"sub_rate"`
	BitsPerMin  float64 `json:"bits_per_min"`
	ChatDensity float64 `json:"chat_density"`
	ComputedAt  string  `json:"computed_at"`
}

func StartMomentumTicker(broadcast chan<- []byte) {
	// run a separate ticker for each streamer
	streamerIDs := []int{1, 3, 4}

	for _, id := range streamerIDs {
		go runTickerForStreamer(id, broadcast)
	}
}

func runTickerForStreamer(streamerID int, broadcast chan<- []byte) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		score, err := Compute(streamerID)
		if err != nil {
			slog.Error("momentum compute failed", "streamer_id", streamerID, "err", err)
			continue
		}

		slog.Info("momentum computed",
			"streamer_id",  streamerID,
			"score",        score.Score,
			"sub_rate",     score.SubRate,
			"bits_per_min", score.BitsPerMin,
			"chat_density", score.ChatDensity,
		)

		payload, err := json.Marshal(score)
		if err != nil {
			continue
		}

		wrapped, _ := json.Marshal(map[string]interface{}{
			"type":    "momentum",
			"payload": score,
			"_meta":   map[string]interface{}{"streamer_id": streamerID},
		})

		db.RDB.Set(context.Background(),
			fmt.Sprintf("streamer:%d:momentum:score", streamerID),
			payload, 10*time.Second)

		broadcast <- wrapped
	}
}


func Compute(streamerID int) (MomentumScore, error) {
	ctx := context.Background()

	prefix := fmt.Sprintf("streamer:%d", streamerID)

	subs, err := db.RDB.Get(ctx, prefix+":counter:sub").Float64()
	if err != nil {
		subs = 0
	}

	bits, err := db.RDB.Get(ctx, prefix+":counter:bits").Float64()
	if err != nil {
		bits = 0
	}

	donations, err := db.RDB.Get(ctx, prefix+":counter:donation").Float64()
	if err != nil {
		donations = 0
	}

	subRate     := subs * 12
	bitsPerMin  := bits * 12
	chatDensity := donations * 12

	score := (subRate * 0.5) + (bitsPerMin * 0.3) + (chatDensity * 0.2)

	db.RDB.Set(ctx, prefix+":counter:sub",      0, 0)
	db.RDB.Set(ctx, prefix+":counter:bits",     0, 0)
	db.RDB.Set(ctx, prefix+":counter:donation", 0, 0)

	return MomentumScore{
		Score:       score,
		SubRate:     subRate,
		BitsPerMin:  bitsPerMin,
		ChatDensity: chatDensity,
		ComputedAt:  time.Now().Format(time.RFC3339),
	}, nil
}
