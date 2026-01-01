package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"real-time-leaderboard/internal/module/leaderboard/application"
	"real-time-leaderboard/internal/module/leaderboard/domain"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a WebSocket client
type Client struct {
	hub                *Hub
	conn               *websocket.Conn
	send               chan []byte
	leaderboardUseCase *application.LeaderboardUseCase
	gameID             string
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients            map[*Client]bool
	broadcast          chan []byte
	register           chan *Client
	unregister         chan *Client
	leaderboardUseCase *application.LeaderboardUseCase
}

// NewHub creates a new hub
func NewHub(leaderboardUseCase *application.LeaderboardUseCase) *Hub {
	return &Hub{
		clients:            make(map[*Client]bool),
		broadcast:          make(chan []byte),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		leaderboardUseCase: leaderboardUseCase,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	ticker := time.NewTicker(5 * time.Second) // Update every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case <-ticker.C:
			// Broadcast leaderboard updates to all clients
			h.broadcastLeaderboard()
		}
	}
}

// broadcastLeaderboard broadcasts leaderboard updates
func (h *Hub) broadcastLeaderboard() {
	// Get global leaderboard
	leaderboard, err := h.leaderboardUseCase.GetGlobalLeaderboard(nil, 10)
	if err != nil {
		log.Printf("Error getting leaderboard: %v", err)
		return
	}

	message, err := json.Marshal(leaderboard)
	if err != nil {
		log.Printf("Error marshaling leaderboard: %v", err)
		return
	}

	// Broadcast to all clients
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// readPump pumps messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		gameID := c.Query("game_id")

		client := &Client{
			hub:                hub,
			conn:               conn,
			send:               make(chan []byte, 256),
			leaderboardUseCase: hub.leaderboardUseCase,
			gameID:             gameID,
		}

		client.hub.register <- client

		// Send initial leaderboard
		var leaderboard *domain.Leaderboard
		var err2 error
		if gameID == "" || gameID == "global" {
			leaderboard, err2 = client.leaderboardUseCase.GetGlobalLeaderboard(c.Request.Context(), 10)
		} else {
			leaderboard, err2 = client.leaderboardUseCase.GetGameLeaderboard(c.Request.Context(), gameID, 10)
		}

		if err2 == nil && leaderboard != nil {
			message, _ := json.Marshal(leaderboard)
			client.send <- message
		}

		go client.writePump()
		go client.readPump()
	}
}
