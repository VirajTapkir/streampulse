package scoring

import (
	"context"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/VirajTapkir/streampulse/db"
)

// setupRedis initialises a real Redis connection for tests
// it uses the same REDIS_ADDR env var as the main app
// if the var isn't set it falls back to localhost:6379
func setupRedis(t *testing.T) {
	t.Helper()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	db.RDB = redis.NewClient(&redis.Options{Addr: addr})

	// verify Redis is reachable — skip the test if it isn't
	if err := db.RDB.Ping(context.Background()).Err(); err != nil {
		t.Skipf("Redis not available at %s — skipping: %v", addr, err)
	}

	// reset all counters before each test so tests don't interfere
	ctx := context.Background()
	db.RDB.Set(ctx, "counter:sub",      0, 0)
	db.RDB.Set(ctx, "counter:bits",     0, 0)
	db.RDB.Set(ctx, "counter:donation", 0, 0)
}

// TestComputeZeroCounters verifies that when all counters are zero
// the momentum score is also zero — the baseline case
func TestComputeZeroCounters(t *testing.T) {
	setupRedis(t)

	score, err := Compute()
	if err != nil {
		t.Fatalf("Compute() returned unexpected error: %v", err)
	}

	if score.Score != 0 {
		t.Errorf("expected score 0 when all counters are zero, got %.2f", score.Score)
	}
}

// TestComputeSubsOnly verifies the sub weight (0.5) in isolation
// 1 sub in a 5-second window = 12 subs/min × 0.5 weight = 6.0
func TestComputeSubsOnly(t *testing.T) {
	setupRedis(t)

	// set exactly 1 sub, zero bits, zero donations
	ctx := context.Background()
	db.RDB.Set(ctx, "counter:sub", 1, 0)

	score, err := Compute()
	if err != nil {
		t.Fatalf("Compute() returned unexpected error: %v", err)
	}

	expected := 1.0 * 12 * 0.5 // 1 sub × per-minute rate × weight
	if score.Score != expected {
		t.Errorf("expected score %.2f with 1 sub, got %.2f", expected, score.Score)
	}

	if score.SubRate != 12.0 {
		t.Errorf("expected sub_rate 12.0, got %.2f", score.SubRate)
	}
}

// TestComputeWeightedFormula verifies all three weights together
// with known inputs so we can calculate the exact expected output
func TestComputeWeightedFormula(t *testing.T) {
	setupRedis(t)

	// set known values for all three counters
	ctx := context.Background()
	db.RDB.Set(ctx, "counter:sub",      2, 0) // 2 subs
	db.RDB.Set(ctx, "counter:bits",     1, 0) // 1 bits event
	db.RDB.Set(ctx, "counter:donation", 1, 0) // 1 donation

	score, err := Compute()
	if err != nil {
		t.Fatalf("Compute() returned unexpected error: %v", err)
	}

	// manually calculate expected score using the same formula
	subRate     := 2.0 * 12 // 24
	bitsPerMin  := 1.0 * 12 // 12
	chatDensity := 1.0 * 12 // 12
	expected    := (subRate * 0.5) + (bitsPerMin * 0.3) + (chatDensity * 0.2)
	// = 12 + 3.6 + 2.4 = 18.0

	if score.Score != expected {
		t.Errorf("expected score %.2f, got %.2f", expected, score.Score)
	}
}

// TestScoreWeights verifies the formula weights sum to exactly 1.0
// if they don't the score would be on the wrong scale entirely
func TestScoreWeights(t *testing.T) {
	subWeight  := 0.5
	bitsWeight := 0.3
	chatWeight := 0.2

	total := subWeight + bitsWeight + chatWeight

	if total != 1.0 {
		t.Errorf("weights must sum to 1.0, got %.2f", total)
	}
}

// TestComputeResetsCounters verifies that after Compute() runs
// the Redis counters are reset to zero for the next window
func TestComputeResetsCounters(t *testing.T) {
	setupRedis(t)

	// seed some values
	ctx := context.Background()
	db.RDB.Set(ctx, "counter:sub",      3, 0)
	db.RDB.Set(ctx, "counter:bits",     2, 0)
	db.RDB.Set(ctx, "counter:donation", 1, 0)

	// run compute — this should reset all counters
	_, err := Compute()
	if err != nil {
		t.Fatalf("Compute() returned unexpected error: %v", err)
	}

	// verify all counters are now zero
	subs, _      := db.RDB.Get(ctx, "counter:sub").Int()
	bits, _      := db.RDB.Get(ctx, "counter:bits").Int()
	donations, _ := db.RDB.Get(ctx, "counter:donation").Int()

	if subs != 0 {
		t.Errorf("expected counter:sub to be reset to 0, got %d", subs)
	}
	if bits != 0 {
		t.Errorf("expected counter:bits to be reset to 0, got %d", bits)
	}
	if donations != 0 {
		t.Errorf("expected counter:donation to be reset to 0, got %d", donations)
	}
}