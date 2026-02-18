package models

import "time"

type BottleEvent struct {
    ID        int64     `json:"id"`
    BottleID  int64     `json:"bottle_id"`
    EventType string    `json:"event_type"`
    Lat       float64   `json:"lat"`
    Lng       float64   `json:"lng"`
    CreatedAt time.Time `json:"created_at"`
}