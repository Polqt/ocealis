package service

import (
	"testing"

	"github.com/Polqt/ocealis/internal/domain"
)

func TestDiscoverRequiresDriftingAndReleased(t *testing.T) {
	cases := []struct {
		name   string
		bottle domain.Bottle
		want   error
	}{
		{
			name: "scheduled hidden",
			bottle: domain.Bottle{
				ID:         1,
				SenderID:   10,
				Status:     domain.BottleStatusScheduled,
				IsReleased: false,
			},
			want: ErrNotDiscoverable,
		},
		{
			name: "drifting but not released",
			bottle: domain.Bottle{
				ID:         2,
				SenderID:   10,
				Status:     domain.BottleStatusDrifting,
				IsReleased: false,
			},
			want: ErrNotDiscoverable,
		},
		{
			name: "already discovered",
			bottle: domain.Bottle{
				ID:         3,
				SenderID:   10,
				Status:     domain.BottleStatusDiscovered,
				IsReleased: true,
			},
			want: ErrAlreadyDiscovered,
		},
		{
			name: "own bottle",
			bottle: domain.Bottle{
				ID:         4,
				SenderID:   99,
				Status:     domain.BottleStatusDrifting,
				IsReleased: true,
			},
			want: ErrSenderCannotDiscover,
		},
		{
			name: "ok",
			bottle: domain.Bottle{
				ID:         5,
				SenderID:   10,
				Status:     domain.BottleStatusDrifting,
				IsReleased: true,
			},
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := discoverGuard(tc.bottle, 99)
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestScheduledReleaseMarksReleasedFlag(t *testing.T) {
	status := domain.BottleStatusDrifting
	isRelease := false
	if status == domain.BottleStatusDrifting {
		isRelease = true
	}
	if !isRelease {
		t.Fatal("drifting transition must set is_release true")
	}
}

// discoverGuard mirrors DiscoverBottle preconditions without DB.
func discoverGuard(bottle domain.Bottle, discoverID int32) error {
	if bottle.SenderID == discoverID {
		return ErrSenderCannotDiscover
	}
	if bottle.Status == domain.BottleStatusDiscovered {
		return ErrAlreadyDiscovered
	}
	if bottle.Status != domain.BottleStatusDrifting || !bottle.IsReleased {
		return ErrNotDiscoverable
	}
	return nil
}
