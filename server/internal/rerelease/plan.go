package rerelease

import (
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Polqt/ocealis/internal/cast"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/geo"
)

// Plan is a Re-release ready to persist at the finder's Shoreline.
type Plan struct {
	Nickname   string
	Lat        float64
	Lng        float64
	VisibleAt  time.Time
	Status     domain.BottleStatus
	IsReleased bool
}

// Prepare validates the finder's Nickname, resolves an Ocean drop, and starts
// a fresh Mystery Delay. Missing location uses the minimal Ocean basin fallback.
func Prepare(nickname string, lat, lng *float64, now time.Time, rng *rand.Rand) (Plan, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return Plan{}, cast.ErrNicknameRequired
	}
	if utf8.RuneCountInString(nickname) > cast.MaxNicknameRunes {
		return Plan{}, cast.ErrNicknameTooLong
	}

	var dropLat, dropLng float64
	if lat == nil || lng == nil {
		fallback := geo.BasinFallback()
		dropLat, dropLng = fallback.Lat, fallback.Lng
	} else {
		dropLat, dropLng = geo.ResolveDrop(*lat, *lng)
	}

	span := cast.MysteryMax - cast.MysteryMin
	offset := cast.MysteryMin + time.Duration(rng.Int63n(int64(span)+1))

	return Plan{
		Nickname:   nickname,
		Lat:        dropLat,
		Lng:        dropLng,
		VisibleAt:  now.Add(offset),
		Status:     domain.BottleStatusMysteryDelay,
		IsReleased: false,
	}, nil
}
