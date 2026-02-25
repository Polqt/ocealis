package ws

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type MessageType string

const (
	MsgBottleDrift      MessageType = "bottle_drift"
	MsgBottleDiscovered MessageType = "bottle_discovered"
	MsgBottleReleased   MessageType = "bottle_released"
)

// Message is the json envelope every connected client receives.
// three.js reads the `type` field to know how to handle the payload
type Message struct {
	Type    MessageType `json:"type"`
	Payload any         `json:"payload"`
}

// DriftPayload is consumed by three.js to update the position of the bottle in the 3D world.
type DriftPayload struct {
	BottleID    int32     `json:"bottle_id"`
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

func NewBroadcaster(hub *Hub, log *zap.Logger) *Broadcaster {
	return &Broadcaster{
		hub: hub,
		log: log,
	}
}

func (b *Broadcaster) BroadcastDrift(payload DriftPayload) {
	// Subscribe watching a specific bottle get full drift updates
	bottleTopic := fmt.Sprintf("bottle:%d", payload.BottleID)

	b.broadcastTopic(bottleTopic, MsgBottleDrift, payload)

	// Subscribers watching a region get a lighter position update without the bottle style or hops,
	// which are only relevant to bottle-specific subscribers.
	regionTopic := regionForCoords(payload.Lat, payload.Lng)
	b.broadcastTopic(regionTopic, MsgBottleDrift, payload)

}

func (b *Broadcaster) BroadcastDiscovered(bottleID int32) {
	topic := fmt.Sprintf("bottle:%d", bottleID)
	b.broadcastTopic(topic, MsgBottleDiscovered, map[string]int32{"bottle_id": bottleID})
}

func (b *Broadcaster) BroadcastReleased(bottleID int32) {
	// New bottles broadcast globally — everyone might want to see a new bottle appear
	b.broadcast(MsgBottleReleased, map[string]int32{"bottle_id": bottleID})
}

func (b *Broadcaster) broadcastTopic(_ string, msgType MessageType, payload any) {
	msg := Message{Type: msgType, Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		b.log.Error("failed to marshal broadcast message", zap.Error(err))
		return
	}
	b.hub.Broadcast(data)
}

func (b *Broadcaster) broadcast(msgType MessageType, payload any) {
	msg := Message{Type: msgType, Payload: payload}
	data, err := json.Marshal(msg)
	if err != nil {
		b.log.Error("failed to marshal ws message", zap.Error(err))
		return
	}
	b.hub.Broadcast(data)
}

// regionForCoords determines the ocean region for given coordinates.
// This is a simple heuristic and can be improved with more accurate geospatial data if needed.
func regionForCoords(lat, lng float64) string {
	switch {
	case lat >= 0 && lng >= -80 && lng <= 0:
		return "region:north_atlantic"
	case lat < 0 && lng >= -60 && lng <= 20:
		return "region:south_atlantic"
	case lat >= 0 && (lng >= 120 || lng <= -120):
		return "region:north_pacific"
	case lat < 0 && (lng >= 150 || lng <= -70):
		return "region:south_pacific"
	case lat >= -60 && lat <= 25 && lng >= 40 && lng <= 120:
		return "region:indian_ocean"
	default:
		return "region:other"
	}
}
