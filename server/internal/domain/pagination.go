package domain

// Cursor is the opaque pagination token passed between pages.
// It wraps the last seen ID so the client never needs to know
// the internal structure, just pass it back on the next request.
type Cursor struct {
	LastID *int32 `json:"last_id,omitempty"`
}

// CursorResult wraps any paginated list with the next cursor.
// When NextCursor.LastID is nil, there are no more pages.
type CursorResult[T any] struct {
	Data       []T     `json:"data"`
	NextCursor *Cursor `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}
