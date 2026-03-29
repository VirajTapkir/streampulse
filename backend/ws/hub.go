package ws

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/VirajTapkir/streampulse/db"
	"github.com/VirajTapkir/streampulse/events"
)

// upgrader converts a plain HTTP connection into a WebSocket connection
// CheckOrigin returns true to allow connections from any origin (fine for dev)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client represents one connected browser tab
type Client struct {
	conn *websocket.Conn // the actual WebSocket connection
	send chan []byte      // buffered channel of messages waiting to be sent
}

// Hub manages all connected clients and broadcasts messages to them
type Hub struct {
	clients    map[*Client]bool // set of currently connected clients
	broadcast  chan []byte       // messages waiting to go out to all clients
	register   chan *Client      // clients waiting to join
	unregister chan *Client      // clients waiting to leave
	mu         sync.Mutex        // protects the clients map from race conditions
}

// NewHub creates and returns a new Hub ready to use
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run is the hub's main loop — it runs forever in its own goroutine
// handling registrations, unregistrations, and broadcasts one at a time
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// a new browser tab connected — add it to the map
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			fmt.Println("client connected, total:", len(h.clients))

		case client := <-h.unregister:
			// a browser tab disconnected — remove it and close its channel
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			fmt.Println("client disconnected, total:", len(h.clients))

		case message := <-h.broadcast:
			// an event arrived — send it to every connected client
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
					// message queued successfully
				default:
					// client's send buffer is full — disconnect it
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// writePump runs in its own goroutine per client
// it reads from the client's send channel and writes to the WebSocket
func (c *Client) writePump() {
	defer c.conn.Close()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return // connection broken, exit the loop
		}
	}
}

// ServeWS handles a new WebSocket connection request
// it upgrades the HTTP connection, registers the client, and starts its pumps
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	// upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	// create the client
	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	// register with the hub
	h.register <- client

	// start the write pump in its own goroutine
	go client.writePump()

	// read pump — keeps the connection alive and handles disconnects
	// gorilla/websocket requires you to read from the connection
	// even if you don't care about incoming messages
	go func() {
		defer func() {
			h.unregister <- client
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break // connection closed
			}
		}
	}()
}

// ProcessEvents reads from the event channel, saves each event to
// PostgreSQL and Redis, then broadcasts it to all connected clients
func (h *Hub) ProcessEvents(eventChan chan events.Event) {
	for evt := range eventChan {
		// 1 — save to PostgreSQL so the earnings history is persistent
		err := saveEarning(evt)
		if err != nil {
			log.Println("failed to save earning:", err)
		}

		// 2 — increment the event counter in Redis for fast reads
		key := fmt.Sprintf("counter:%s", evt.Type)
		db.RDB.Incr(context.Background(), key)

		// 3 — marshal the event to JSON and broadcast to all clients
		payload, err := json.Marshal(map[string]interface{}{
			"type":      evt.Type,
			"username":  evt.Username,
			"amount":    evt.Amount,
			"timestamp": evt.Timestamp.Format(time.RFC3339),
		})
		if err != nil {
			continue
		}

		h.broadcast <- payload
	}
}

// saveEarning inserts one event into the earnings table
// it always uses streamer_id = 1 for now (our test streamer)
func saveEarning(evt events.Event) error {
	_, err := db.DB.Exec(
		`INSERT INTO earnings (streamer_id, event_type, amount) VALUES ($1, $2, $3)`,
		1, evt.Type, evt.Amount,
	)
	return err
}

// GetCounters handles GET /api/counters
// returns the live event counts stored in Redis
func GetCounters(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// read all three counters from Redis
	subs, _ := db.RDB.Get(ctx, "counter:sub").Int()
	bits, _ := db.RDB.Get(ctx, "counter:bits").Int()
	donations, _ := db.RDB.Get(ctx, "counter:donation").Int()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"subs":      subs,
		"bits":      bits,
		"donations": donations,
	})
}

// GetDB exposes the sql.DB for use in handlers that need raw SQL access
func GetDB() *sql.DB {
	return db.DB
}

// GetBroadcast returns the hub's broadcast channel
// so external packages like scoring can send messages to all clients
func (h *Hub) GetBroadcast() chan<- []byte {
	return h.broadcast
}