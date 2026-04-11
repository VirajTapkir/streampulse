package ws

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	// "time"

	"github.com/gorilla/websocket"
	"github.com/VirajTapkir/streampulse/db"
	"github.com/VirajTapkir/streampulse/events"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn       *websocket.Conn
	send       chan []byte
	streamerID int 
}
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			// structured log — shows exactly how many clients are connected
			slog.Info("client connected", "total_clients", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			slog.Info("client disconnected", "total_clients", len(h.clients))

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				// only send to clients watching this streamer
				// parse streamer_id from the message to filter correctly
				var payload map[string]interface{}
				if err := json.Unmarshal(message, &payload); err == nil {
					if meta, ok := payload["_meta"].(map[string]interface{}); ok {
						if sid, ok := meta["streamer_id"].(float64); ok {
							if client.streamerID != int(sid) {
								continue // skip clients watching a different streamer
							}
						}
					}
				}
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					slog.Warn("client removed — send buffer full")
				}
			}
			h.mu.Unlock()

		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			slog.Error("websocket write failed", "err", err)
			return
		}
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "err", err)
		return
	}

	streamerID := 1
	if sid := r.URL.Query().Get("streamer_id"); sid != "" {
		if id, err := fmt.Sscanf(sid, "%d", &streamerID); err != nil || id == 0 {
			streamerID = 1
		}
	}

	client := &Client{
		conn:       conn,
		send:       make(chan []byte, 256),
		streamerID: streamerID,
	}

	slog.Info("client connected", "streamer_id", streamerID)

	h.register <- client
	go client.writePump()

	go func() {
		defer func() {
			h.unregister <- client
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

func (h *Hub) ProcessEvents(eventChan chan events.Event) {
	for evt := range eventChan {
		if err := saveEarning(evt); err != nil {
			slog.Error("failed to save earning", "event_type", evt.Type, "err", err)
		}

		// namespace Redis keys by streamer ID so each streamer's
		// counters are completely independent
		key := fmt.Sprintf("streamer:%d:counter:%s", evt.StreamerID, evt.Type)
		db.RDB.Incr(context.Background(), key)


		slog.Info("event processed",
			"type",     evt.Type,
			"username", evt.Username,
			"amount",   evt.Amount,
		)

		payload, err := json.Marshal(evt.TwitchPayload)

		if err != nil {
			slog.Error("failed to marshal event", "err", err)
			continue
		}

		h.broadcast <- payload
	}
}

func saveEarning(evt events.Event) error {
	_, err := db.DB.Exec(
		`INSERT INTO earnings (streamer_id, event_type, amount) VALUES ($1, $2, $3)`,
		evt.StreamerID, evt.Type, evt.Amount,
	)
	return err
}


func GetCounters(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// read streamer_id from query param — default to 1
	streamerID := "1"
	if sid := r.URL.Query().Get("streamer_id"); sid != "" {
		streamerID = sid
	}

	subs, _      := db.RDB.Get(ctx, fmt.Sprintf("streamer:%s:counter:sub", streamerID)).Int()
	bits, _      := db.RDB.Get(ctx, fmt.Sprintf("streamer:%s:counter:bits", streamerID)).Int()
	donations, _ := db.RDB.Get(ctx, fmt.Sprintf("streamer:%s:counter:donation", streamerID)).Int()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"subs":      subs,
		"bits":      bits,
		"donations": donations,
	})
}


func GetDB() *sql.DB {
	return db.DB
}

func (h *Hub) GetBroadcast() chan<- []byte {
	return h.broadcast
}