package domain

import "time"

// BottleStatus is the Bottle life-cycle status (CONTEXT.md / PRD).
type BottleStatus string

const (
	// BottleStatusDrifting — visible Cork in the Ocean.
	BottleStatusDrifting BottleStatus = "drifting"
	// BottleStatusMysteryDelay — invisible after Cast/Re-release until VisibleAt.
	// Wire/DB value still "scheduled" until column migration (deferred).
	BottleStatusMysteryDelay BottleStatus = "scheduled"
	// BottleStatusSunk — left the world after Sink (stub until issue 07).
	BottleStatusSunk BottleStatus = "sunk"
	// BottleStatusClaimed — legacy claim status. Open must not set this (issue 03).
	// Wire value "discovered". Prefer Open (read-only) + Stamp/Re-release.
	BottleStatusClaimed BottleStatus = "discovered"
)

// Deprecated aliases — prefer glossary names above.
const (
	BottleStatusDiscovered = BottleStatusClaimed
	BottleStatusScheduled  = BottleStatusMysteryDelay
	BottleStatusReleased   BottleStatus = "released" // unused life status; Cast is an event
)

type Bottle struct {
	ID          int32        `json:"id"`
	SenderID    int32        `json:"sender_id,omitempty"` // legacy account FK; Nickname is product identity
	Nickname    string       `json:"nickname"`
	MessageText string       `json:"message_text"`
	BottleStyle int32        `json:"bottle_style"`
	StartLat    float64      `json:"start_lat"`
	StartLng    float64      `json:"start_lng"`
	CurrentLat  float64      `json:"current_lat"`
	CurrentLng  float64      `json:"current_lng"`
	Hops        int32        `json:"hops"`
	// VisibleAt is Mystery Delay end — Cork appears after this instant.
	// Maps from DB scheduled_release until rename.
	VisibleAt time.Time `json:"visible_at"`
	// ScheduledRelease kept for older clients; same instant as VisibleAt.
	ScheduledRelease time.Time    `json:"scheduled_release,omitempty"`
	IsReleased       bool         `json:"is_released"`
	Status           BottleStatus `json:"status"`
	CreatedAt        time.Time    `json:"created_at"`
}

type Journey struct {
	Bottle *Bottle       `json:"bottle"`
	Events []BottleEvent `json:"events"`
}
