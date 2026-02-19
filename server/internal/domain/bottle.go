package domain

import "time"

type BottleStatus string

const (
	BottleStatusDrifting   BottleStatus = "drifting"
	BottleStatusDiscovered BottleStatus = "discovered"
	BottleStatusReleased   BottleStatus = "released"
)

type Bottle struct {
	ID               int32        `json:"id"`
	SenderID         int32        `json:"sender_id"`
	MessageText      string       `json:"message_text"`
	BottleStyle      int32        `json:"bottle_style"`
	StartLat         float64      `json:"start_lat"`
	StartLng         float64      `json:"start_lng"`
	CurrentLat       float64      `json:"current_lat"`
	CurrentLng       float64      `json:"current_lng"`
	Hops             int32        `json:"hops"`
	ScheduledRelease time.Time    `json:"scheduled_release"`
	IsReleased       bool         `json:"is_released"`
	Status           BottleStatus `json:"status"`
	CreatedAt        time.Time    `json:"created_at"`
}

type Journey struct {
	Bottle *Bottle `json:"bottle"`
	Event []EventType `json:"events"`
}
