package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/VirajTapkir/streampulse/api"
	"github.com/VirajTapkir/streampulse/db"
	"github.com/VirajTapkir/streampulse/events"
	"github.com/VirajTapkir/streampulse/middleware"
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

	// start the momentum score ticker
	scoring.StartMomentumTicker(hub.GetBroadcast())

	// set up a ServeMux so we can wrap ALL routes with CORS at once
	mux := http.NewServeMux()

	// REST API routes
	mux.HandleFunc("/api/streamers", api.GetStreamers)
	mux.HandleFunc("/api/earnings",  api.GetEarnings)
	mux.HandleFunc("/api/counters",  ws.GetCounters)
	mux.HandleFunc("/api/momentum",  api.GetMomentum)

	// health check endpoint — AWS ECS uses this to verify the container is alive
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// WebSocket route
	mux.HandleFunc("/ws", hub.ServeWS)

	// wrap the entire mux with CORS middleware
	// every single request now gets CORS headers automatically
	handler := middleware.CORS(mux)

	// create the HTTP server as a variable so we can shut it down gracefully
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,

		// timeouts prevent slow or malicious clients from
		// holding connections open and exhausting server resources
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// launch the server in a goroutine so it doesn't block
	// the graceful shutdown logic below
	go func() {
		fmt.Println("StreamPulse backend running on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// graceful shutdown setup
	// make a channel that receives OS signals
	quit := make(chan os.Signal, 1)

	// tell Go to forward Ctrl+C (SIGINT) and system stop (SIGTERM) into that channel
	// SIGTERM is what AWS ECS sends when it wants to stop your container
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// block here doing nothing until a signal arrives
	<-quit
	fmt.Println("shutdown signal received — draining connections...")

	// give in-flight requests up to 10 seconds to finish
	// after that, force close everything
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	fmt.Println("StreamPulse stopped cleanly")
}