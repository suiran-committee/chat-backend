package ws

import (
	"context"
	"log"
	"sync"

	"chat-backend/internal/model"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
	in      chan model.Message
}

func NewHub(buf int) *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]struct{}),
		in:      make(chan model.Message, buf),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-h.in:
			h.broadcast(m)
		}
	}
}

func (h *Hub) Register(c *websocket.Conn)   { h.mu.Lock(); h.clients[c] = struct{}{}; h.mu.Unlock() }
func (h *Hub) Unregister(c *websocket.Conn) { h.mu.Lock(); delete(h.clients, c); h.mu.Unlock() }
func (h *Hub) Send(m model.Message)         { h.in <- m }

func (h *Hub) broadcast(m model.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if err := c.WriteJSON(m); err != nil {
			log.Printf("[ws] write err: %v", err)
			c.Close()
			delete(h.clients, c)
		}
	}
}
