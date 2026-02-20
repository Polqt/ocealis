package ws

import (
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

type MessageType string

const (
	MsgBottleDrift      MessageType = "bottle_drift"
	MSgBottleDiscovered MessageType = "bottle_discovered"
	MSGBottleReleased   MessageType = "bottle_released"
)

// Message is the json envelope every connected client receives.
// three.js reads the `type` field to knwow how to handle the payload
type Message struct {
	Type    MessageType `json:"type"`
	Payload any         `json:"payload"`
}

// Drift payload is consumed by three.js to update the position of the bottle in the 3D world.
type DriftPayload struct {
	BottleID    string    `json:"bottle_id"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Hops        int32     `json:"hops"`
	BottleStyle int32     `json:"bottle_style"`
	Timestamp   time.Time `json:"timestamp"`
}

// Broadcaster wraps hub with typed, domain-specific messages and payloads, so that other parts of the server can broadcast messages without worrying about the underlying WebSocket implementation.
// Services call broadcaster to broadcast messages to all connected clients, and the broadcaster translates them into the appropriate format for the hub to send. This separation of concerns allows for cleaner code and easier maintenance.
type Broadcaster struct {
	hub *Hub
	log *zap.Logger
}

func NewBroadcasteR(hub *Hub, log *zap.Logger) *Broadcaster {
	return &Broadcaster{
		hub: hub,
		log: log,
	}
}

func (b *Broadcaster) BroadcastDrift(payload DriftPayload) {
	b.broadcast(MsgBottleDrift, payload)
}

func (b *Broadcaster) BroadcastDiscovered(bottleID int32) {
	b.broadcast(MSgBottleDiscovered, map[string]int32{"bottle_id": bottleID})
}

func (b *Broadcaster) BroadcastReleased(bottleID int32) {
	b.broadcast(MSGBottleReleased, map[string]int32{"bottle_id": bottleID})
}

func (b *Broadcaster) broadcast(msgType MessageType, payload any) {
	msg := Message{Type: msgType, Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		b.log.Error("failed to marshal broadcast message", zap.Error(err))
		return
	}
	b.hub.Broadcast(data)
}
