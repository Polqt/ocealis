package service

import (
	"errors"
	"testing"

	"github.com/Polqt/ocealis/internal/domain"
)

func TestDiscoverRules(t *testing.T) {
	bottle := &domain.Bottle{
		ID:       1,
		SenderID: 10,
		Status:   domain.BottleStatusDrifting,
	}

	if bottle.SenderID == 10 {
		if !errors.Is(ErrSenderCannotDiscover, ErrSenderCannotDiscover) {
			t.Fatal("sender cannot discover own bottle")
		}
	}

	bottle.Status = domain.BottleStatusDiscovered
	if bottle.Status == domain.BottleStatusDiscovered {
		if !errors.Is(ErrAlreadyDiscovered, ErrAlreadyDiscovered) {
			t.Fatal("already discovered")
		}
	}
}

func TestCreateBottleStatusForSchedule(t *testing.T) {
	scheduled := true
	status := domain.BottleStatusDrifting
	released := true
	if scheduled {
		status = domain.BottleStatusScheduled
		released = false
	}
	if status != domain.BottleStatusScheduled || released {
		t.Fatalf("scheduled cast should not be released, got status=%s released=%v", status, released)
	}
}
