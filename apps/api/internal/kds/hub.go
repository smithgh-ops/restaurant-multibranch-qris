package kds

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second
	maxMessageSize = 512
)

type realtimeClient struct {
	conn      *websocket.Conn
	send      chan []byte
	closeOnce sync.Once
}

func (c *realtimeClient) close() {
	c.closeOnce.Do(func() {
		close(c.send)
		if c.conn != nil {
			_ = c.conn.Close()
		}
	})
}

// Hub manages websocket clients per branch and broadcasts realtime KDS events.
type Hub struct {
	mu      sync.RWMutex
	clients map[uint64]map[*realtimeClient]struct{}
}

// NewHub creates a new KDS realtime hub.
func NewHub() *Hub {
	return &Hub{clients: make(map[uint64]map[*realtimeClient]struct{})}
}

// ServeConnection registers a websocket connection for a branch.
func (h *Hub) ServeConnection(branchID uint64, conn *websocket.Conn) {
	client := &realtimeClient{
		conn: conn,
		send: make(chan []byte, 16),
	}
	h.addClient(branchID, client)
	go client.writePump(h, branchID)
	go client.readPump(h, branchID)
}

// Broadcast sends a payload to all clients subscribed to a branch.
func (h *Hub) Broadcast(branchID uint64, payload any) error {
	message, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	h.mu.RLock()
	branchClients := h.clients[branchID]
	clients := make([]*realtimeClient, 0, len(branchClients))
	for client := range branchClients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			h.removeClient(branchID, client)
		}
	}
	return nil
}

func (h *Hub) addClient(branchID uint64, client *realtimeClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[branchID]; !ok {
		h.clients[branchID] = make(map[*realtimeClient]struct{})
	}
	h.clients[branchID][client] = struct{}{}
}

func (h *Hub) removeClient(branchID uint64, client *realtimeClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	branchClients, ok := h.clients[branchID]
	if !ok {
		client.close()
		return
	}
	if _, exists := branchClients[client]; exists {
		delete(branchClients, client)
	}
	if len(branchClients) == 0 {
		delete(h.clients, branchID)
	}
	client.close()
}

func (c *realtimeClient) readPump(h *Hub, branchID uint64) {
	defer h.removeClient(branchID, c)
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *realtimeClient) writePump(h *Hub, branchID uint64) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		h.removeClient(branchID, c)
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
