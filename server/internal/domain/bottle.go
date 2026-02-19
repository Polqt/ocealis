package models

import "time"

type Bottle struct {
	ID               int64     `json:"id"`
	SenderID         int64     `json:"sender_id"`
	MessageText      string    `json:"message_text"`
	BottleStyle      int       `json:"bottle_style"`
	StartLat         float64   `json:"start_lat"`
	StartLng         float64   `json:"start_lng"`
	CurrentLat       float64   `json:"current_lat"`
	CurrentLng       float64   `json:"current_lng"`
	Hops             int       `json:"hops"`
	ScheduledRelease time.Time `json:"scheduled_release"`
	IsReleased       bool      `json:"is_released"`
	CreatedAt        time.Time `json:"created_at"`
}
