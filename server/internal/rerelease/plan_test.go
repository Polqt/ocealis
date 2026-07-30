package rerelease_test

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/geo"
	"github.com/Polqt/ocealis/internal/rerelease"
)

func TestPrepareRequiresNickname(t *testing.T) {
	now := time.Date(2026, 7, 30, 1, 0, 0, 0, time.UTC)
	lat, lng := 30.0, -140.0

	if _, err := rerelease.Prepare("  ", &lat, &lng, now, rand.New(rand.NewSource(1))); err == nil {
		t.Fatal("Re-release must require Nickname")
	}
	if _, err := rerelease.Prepare(strings.Repeat("n", 25), &lat, &lng, now, rand.New(rand.NewSource(1))); err == nil {
		t.Fatal("Re-release must reject Nickname over 24 characters")
	}
}

func TestPrepareSnapsToShorelineAndAppliesMysteryDelay(t *testing.T) {
	now := time.Date(2026, 7, 30, 1, 0, 0, 0, time.UTC)
	lat, lng := 39.0997, -94.5786 // inland Kansas City

	plan, err := rerelease.Prepare("  finder  ", &lat, &lng, now, rand.New(rand.NewSource(42)))
	if err != nil {
		t.Fatal(err)
	}

	if plan.Nickname != "finder" {
		t.Fatalf("Nickname=%q want trimmed finder", plan.Nickname)
	}
	if plan.Status != domain.BottleStatusMysteryDelay || plan.IsReleased {
		t.Fatalf("Re-release must enter invisible Mystery Delay; got status=%q released=%v", plan.Status, plan.IsReleased)
	}
	if plan.VisibleAt.Before(now.Add(15*time.Minute)) || plan.VisibleAt.After(now.Add(30*time.Minute)) {
		t.Fatalf("visible_at=%v outside 15–30 minute Mystery Delay", plan.VisibleAt)
	}
	if geo.IsLand(plan.Lat, plan.Lng) {
		t.Fatalf("Re-release drop must be Ocean, got %v,%v", plan.Lat, plan.Lng)
	}
	if plan.Lat == lat && plan.Lng == lng {
		t.Fatal("inland Re-release must relocate to finder Shoreline")
	}
}

func TestPrepareMissingGeoUsesOceanBasin(t *testing.T) {
	now := time.Date(2026, 7, 30, 1, 0, 0, 0, time.UTC)
	plan, err := rerelease.Prepare("finder", nil, nil, now, rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatal(err)
	}

	fallback := geo.BasinFallback()
	if plan.Lat != fallback.Lat || plan.Lng != fallback.Lng {
		t.Fatalf("want Ocean basin fallback %+v, got %v,%v", fallback, plan.Lat, plan.Lng)
	}
}
