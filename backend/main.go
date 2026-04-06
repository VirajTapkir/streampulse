package main

import (
	"context"
	"log/slog"
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
	// configure slog to write structured logs to stdout
	// this is what AWS CloudWatch will capture in production
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // show INFO and above — debug logs are hidden
	}))
	slog.SetDefault(logger)

	// load .env
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found — using environment variables")
	}

	// connect to PostgreSQL
	if err := db.InitPostgres(); err != nil {
		slog.Error("postgres init failed", "err", err)
		os.Exit(1)
	}

	// connect to Redis
	if err := db.InitRedis(); err != nil {
		slog.Error("redis init failed", "err", err)
		os.Exit(1)
	}

	// create the WebSocket hub and start its main loop
	hub := ws.NewHub()
	go hub.Run()

	// start the mock event queue
	eventChan := events.StartEventQueue()

	// process events
	go hub.ProcessEvents(eventChan)

	// start momentum ticker
	scoring.StartMomentumTicker(hub.GetBroadcast())

	// set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/streamers", api.GetStreamers)
	mux.HandleFunc("/api/earnings",  api.GetEarnings)
	mux.HandleFunc("/api/counters",  ws.GetCounters)
	mux.HandleFunc("/api/momentum",  api.GetMomentum)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ws", hub.ServeWS)

	handler := middleware.CORS(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server started", "addr", ":8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received — draining connections")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "err", err)
		os.Exit(1)
	}

	slog.Info("StreamPulse stopped cleanly")
}