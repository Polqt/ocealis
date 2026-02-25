package ws

import "sync"

// Hub maintains the set of all active WebSocket connections and broadcasts messages to the connections.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	topics  map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		topics:  make(map[string]map[*Client]struct{}),
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

	// Remove from global client list
	delete(h.clients, c)

	// Remove from all topic subscriptions
	for topic, subs := range h.topics {
		delete(subs, c)
		// Clean up empty topic entries
		if len(subs) == 0 {
			delete(h.topics, topic)
		}
	}
}

// Subscribe adds a client to a topic's subscriber list.
func (h *Hub) Subscribe(c *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.topics[topic] == nil {
		h.topics[topic] = make(map[*Client]struct{})
	}

	h.topics[topic][c] = struct{}{}
}

// Unsubscribe removes a client from a topic's subscriber list.
func (h *Hub) Unsubscribe(c *Client, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.topics[topic]; ok {
		delete(subs, c)
		if len(subs) == 0 {
			delete(h.topics, topic)
		}
	}
}

// Broadcast sends to all connected clients regardless of subscriptions.
// Use for global events like storms.
func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
		}
	}
}

// Broadcast Topic sends only to clients subscribed to a specific topic.
func (h *Hub) BroadcastTopic(topic string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs, ok := h.topics[topic]
	if !ok {
		return // No subscribers for this topic
	}

	for c := range subs {
		select {
		case c.send <- msg:
		default:
		}
	}
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
