package discovery

import (
	"time"

	"github.com/Polqt/ocealis/internal/domain"
)

// SeedStyle marks a Seed Bottle via bottle_style (no schema column yet).
// Issue 07 may promote to is_seed + Sink exemption.
const SeedStyle int32 = 9

// Seeds returns always-visible Seed Bottles so the Ocean is never empty.
// Negative IDs avoid colliding with Postgres serial until seeds land in DB.
func Seeds() []domain.Bottle {
	past := time.Now().Add(-24 * time.Hour)
	return []domain.Bottle{
		seed(-1, "Ocealis", "The ocean keeps what you let go.", 30.0, -140.0, past),
		seed(-2, "Ocealis", "A cork drifts farther than a plan.", 0.0, -30.0, past),
		seed(-3, "Ocealis", "Leave a Message. Keep no map pin.", -20.0, 160.0, past),
		seed(-4, "Ocealis", "Shore to shore, no names required.", 45.0, -20.0, past),
		seed(-5, "Ocealis", "Open a Cork. Stamp a Journey. Re-release.", 15.0, 120.0, past),
	}
}

func seed(id int32, nick, msg string, lat, lng float64, visibleAt time.Time) domain.Bottle {
	return domain.Bottle{
		ID:          id,
		Nickname:    nick,
		MessageText: msg,
		BottleStyle: SeedStyle,
		StartLat:    lat,
		StartLng:    lng,
		CurrentLat:  lat,
		CurrentLng:  lng,
		Status:      domain.BottleStatusDrifting,
		IsReleased:  true,
		VisibleAt:   visibleAt,
		CreatedAt:   visibleAt,
	}
}
