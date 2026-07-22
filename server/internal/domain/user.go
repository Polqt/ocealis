package domain

import "time"

// User is quarantined — accounts/JWT are not product v1 (PRD US28).
// Nickname on Cast/Re-release is metadata, not this type.

type User struct {
	ID        int32     `json:"id"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
