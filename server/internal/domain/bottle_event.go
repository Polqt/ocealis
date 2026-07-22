package domain

import "time"

// EventType is a Journey event (CONTEXT.md).
type EventType string

const (
	// EventTypeCast — Visitor Cast a Bottle into the Ocean.
	// Wire value still "released" until event_type migration (deferred).
	EventTypeCast EventType = "released"
	EventTypeDrift EventType = "drift"
	// EventTypeStamp — passport seal/note on Journey (stub until issue 04).
	EventTypeStamp EventType = "stamp"
	// EventTypeReReleased — finder Re-release from their Shoreline.
	EventTypeReReleased EventType = "re_released"
	// EventTypeSink — Bottle left the world (stub until issue 07).
	EventTypeSink EventType = "sink"
	// EventTypeOpenedLegacy — old claim event. Open is read-only (issue 03);
	// wire value "discovered".
	EventTypeOpenedLegacy EventType = "discovered"
)

// Deprecated aliases — prefer glossary names above.
const (
	EventTypeReleased   = EventTypeCast
	EventTypeDiscovered = EventTypeOpenedLegacy
)

type BottleEvent struct {
	ID        int32     `json:"id"`
	BottleID  int32     `json:"bottle_id"`
	EventType EventType `json:"event_type"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	CreatedAt time.Time `json:"created_at"`
}
