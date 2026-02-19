package domain

import "time"

type EventType string

const (
	EventTypeReleased   EventType = "released"
	EventTypeDrift      EventType = "drift"
	EventTypeDiscovered EventType = "discovered"
	EventTypeReReleased EventType = "re_released"
)

type BottleEvent struct {
	ID        int32     `json:"id"`
	BottleID  int32     `json:"bottle_id"`
	EventType EventType `json:"event_type"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	CreatedAt time.Time `json:"created_at"`
}
