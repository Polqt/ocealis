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

// Broadcast sends msg to every connected client.
// It takes a snapshot under the read lock so that writes to individual
// send channels never block while the lock is held.
func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	// Send to the snapshot — no lock held, no data race.
	var stale []*Client
	for _, c := range clients {
		select {
		case c.send <- msg:
		default:
			// Send buffer full: mark client for removal.
			stale = append(stale, c)
		}
	}

	// Evict unresponsive clients outside the send loop.
	for _, c := range stale {
		h.Unregister(c)
	}
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
