package ws

import "sync"

// Hub maintains the set of all active WebSocket connections and broadcasts messages to the connections.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
}

// Broadcast sends a message to every connected clients.
func (h *Hub) Broadcast(msg []byte) {
	// Take a snapshot of current clients under read lock,
	// so that we can send messages without holding the lock,
	// allowing new clients to register or unregister concurrently.
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	var showClients []*Client

	for c := range h.clients {
		select {
		case c.send <- msg:
			// Message sent successfully
		default:
			// If the client's send channel is full, we can choose to drop the message or handle it as needed.
			// For now, we'll just skip sending to this client.
			showClients = append(showClients, c)
		}
	}

	// Unregister clients that are not responsive (send channel is full).
	for _, c := range showClients {
		// Optionally log slow/stuck clients before unregistering.
		// Example (requires a logger or the standard log package):
		// log.Printf("ws hub: unregistering slow client %v: send buffer full", c)
		h.Unregister(c)
	}
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
