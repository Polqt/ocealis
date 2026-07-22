package cast

import (
	"errors"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/geo"
	"github.com/Polqt/ocealis/util"
)

const (
	MaxNicknameRunes = 24
	MaxMessageRunes  = 500
	MysteryMin       = 15 * time.Minute
	MysteryMax       = 30 * time.Minute
)

var (
	ErrNicknameRequired = errors.New("nickname required")
	ErrNicknameTooLong  = errors.New("nickname must be ≤24 characters")
	ErrMessageRequired  = errors.New("message required")
	ErrMessageTooLong   = errors.New("message must be ≤500 characters")
)

// Plan is the Cast drop ready to persist — Ocean coords + Mystery Delay.
type Plan struct {
	Nickname    string
	MessageText string
	Lat         float64
	Lng         float64
	VisibleAt   time.Time
	Status      domain.BottleStatus
	IsReleased  bool
}

// Prepare validates Cast inputs, snaps inland to Shoreline, applies Mystery Delay.
// lat/lng nil → BasinFallback (denied/missing geo).
func Prepare(nickname, message string, lat, lng *float64, now time.Time, rng *rand.Rand) (Plan, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" {
		return Plan{}, ErrNicknameRequired
	}
	if utf8.RuneCountInString(nickname) > MaxNicknameRunes {
		return Plan{}, ErrNicknameTooLong
	}

	message = util.SanitizeMessage(message)
	message = strings.TrimSpace(message)
	if message == "" {
		return Plan{}, ErrMessageRequired
	}
	if utf8.RuneCountInString(message) > MaxMessageRunes {
		return Plan{}, ErrMessageTooLong
	}

	var dropLat, dropLng float64
	if lat == nil || lng == nil {
		fb := geo.BasinFallback()
		dropLat, dropLng = fb.Lat, fb.Lng
	} else {
		dropLat, dropLng = geo.ResolveDrop(*lat, *lng)
	}

	span := MysteryMax - MysteryMin
	offset := MysteryMin + time.Duration(rng.Int63n(int64(span)+1))
	visibleAt := now.Add(offset)

	return Plan{
		Nickname:    nickname,
		MessageText: message,
		Lat:         dropLat,
		Lng:         dropLng,
		VisibleAt:   visibleAt,
		Status:      domain.BottleStatusMysteryDelay,
		IsReleased:  false,
	}, nil
}
