package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/VirajTapkir/streampulse/api"
	"github.com/VirajTapkir/streampulse/db"
	"github.com/VirajTapkir/streampulse/events"
	"github.com/VirajTapkir/streampulse/scoring"
	"github.com/VirajTapkir/streampulse/ws"
)

func main() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// connect to PostgreSQL
	if err := db.InitPostgres(); err != nil {
		log.Fatalf("Postgres error: %v", err)
	}

	// connect to Redis
	if err := db.InitRedis(); err != nil {
		log.Fatalf("Redis error: %v", err)
	}

	// create the WebSocket hub and start its main loop
	hub := ws.NewHub()
	go hub.Run()

	// start the mock event queue
	eventChan := events.StartEventQueue()

	// process events — saves to DB, updates Redis, broadcasts to clients
	go hub.ProcessEvents(eventChan)

	// start the momentum score ticker — recomputes every 5 seconds
	// we pass hub.Broadcast so the scorer can send directly to all clients
	scoring.StartMomentumTicker(hub.GetBroadcast())

	// REST API routes
	http.HandleFunc("/api/streamers", api.GetStreamers)
	http.HandleFunc("/api/earnings", api.GetEarnings)
	http.HandleFunc("/api/counters", ws.GetCounters)
	http.HandleFunc("/api/momentum", api.GetMomentum)

	// WebSocket route
	http.HandleFunc("/ws", hub.ServeWS)

	fmt.Println("StreamPulse backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}