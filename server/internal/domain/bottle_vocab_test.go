package domain_test

import (
	"testing"

	"github.com/Polqt/ocealis/internal/domain"
)

func TestBottleLifeStatusesUseGlossaryWireValues(t *testing.T) {
	// Product names map to wire/DB strings (rename deferred where noted).
	cases := []struct {
		status domain.BottleStatus
		wire   string
	}{
		{domain.BottleStatusDrifting, "drifting"},
		{domain.BottleStatusMysteryDelay, "scheduled"}, // Mystery Delay
		{domain.BottleStatusSunk, "sunk"},              // Sink stub
		{domain.BottleStatusClaimed, "discovered"},     // legacy claim
	}
	for _, tc := range cases {
		if string(tc.status) != tc.wire {
			t.Fatalf("%q wire = %q, want %q", tc.status, tc.status, tc.wire)
		}
	}
}

func TestJourneyEventTypesCoverCastStampReReleaseSink(t *testing.T) {
	cases := []struct {
		event domain.EventType
		wire  string
	}{
		{domain.EventTypeCast, "released"}, // Cast; wire rename deferred
		{domain.EventTypeDrift, "drift"},
		{domain.EventTypeStamp, "stamp"},
		{domain.EventTypeReReleased, "re_released"},
		{domain.EventTypeSink, "sink"},
	}
	for _, tc := range cases {
		if string(tc.event) != tc.wire {
			t.Fatalf("%q wire = %q, want %q", tc.event, tc.event, tc.wire)
		}
	}
}
