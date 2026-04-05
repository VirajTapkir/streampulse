package scoring

import (
	"context"
	"encoding/json"
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
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			score, err := Compute()
			if err != nil {
				slog.Error("momentum compute failed", "err", err)
				continue
			}

			slog.Info("momentum computed",
				"score",        score.Score,
				"sub_rate",     score.SubRate,
				"bits_per_min", score.BitsPerMin,
				"chat_density", score.ChatDensity,
			)

			payload, err := json.Marshal(score)
			if err != nil {
				slog.Error("failed to marshal momentum score", "err", err)
				continue
			}

			wrapped, _ := json.Marshal(map[string]interface{}{
				"type":    "momentum",
				"payload": score,
			})

			db.RDB.Set(context.Background(), "momentum:score", payload, 10*time.Second)
			broadcast <- wrapped
		}
	}()
}

func Compute() (MomentumScore, error) {
	ctx := context.Background()

	subs, err := db.RDB.Get(ctx, "counter:sub").Float64()
	if err != nil {
		subs = 0
	}

	bits, err := db.RDB.Get(ctx, "counter:bits").Float64()
	if err != nil {
		bits = 0
	}

	donations, err := db.RDB.Get(ctx, "counter:donation").Float64()
	if err != nil {
		donations = 0
	}

	subRate     := subs * 12
	bitsPerMin  := bits * 12
	chatDensity := donations * 12

	score := (subRate * 0.5) + (bitsPerMin * 0.3) + (chatDensity * 0.2)

	db.RDB.Set(ctx, "counter:sub",      0, 0)
	db.RDB.Set(ctx, "counter:bits",     0, 0)
	db.RDB.Set(ctx, "counter:donation", 0, 0)

	return MomentumScore{
		Score:       score,
		SubRate:     subRate,
		BitsPerMin:  bitsPerMin,
		ChatDensity: chatDensity,
		ComputedAt:  time.Now().Format(time.RFC3339),
	}, nil
}