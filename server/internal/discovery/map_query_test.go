package discovery_test

import (
	"testing"
	"time"

	"github.com/Polqt/ocealis/internal/discovery"
	"github.com/Polqt/ocealis/internal/domain"
)

func pacificVP() discovery.Viewport {
	return discovery.Viewport{MinLat: 20, MaxLat: 40, MinLng: -160, MaxLng: -120}
}

func TestFarZoomReturnsHeatNotCorks(t *testing.T) {
	now := time.Now()
	bottles := []domain.Bottle{
		{ID: 1, Status: domain.BottleStatusDrifting, IsReleased: true, VisibleAt: now.Add(-time.Hour), CurrentLat: 30, CurrentLng: -140},
		{ID: 2, Status: domain.BottleStatusDrifting, IsReleased: true, VisibleAt: now.Add(-time.Hour), CurrentLat: 31, CurrentLng: -141},
	}
	out := discovery.QueryOcean(3, pacificVP(), bottles)
	if out.Mode != "heat" {
		t.Fatalf("far zoom want heat, got %q", out.Mode)
	}
	if len(out.Corks) != 0 {
		t.Fatalf("far zoom must not return corks; got %d", len(out.Corks))
	}
	if len(out.Heat) == 0 {
		t.Fatal("far zoom want heat cells")
	}
}

func TestNearZoomReturnsCorks(t *testing.T) {
	now := time.Now()
	bottles := []domain.Bottle{
		{ID: 1, Status: domain.BottleStatusDrifting, IsReleased: true, VisibleAt: now.Add(-time.Hour), CurrentLat: 30, CurrentLng: -140},
	}
	out := discovery.QueryOcean(6, pacificVP(), bottles)
	if out.Mode != "corks" {
		t.Fatalf("near zoom want corks, got %q", out.Mode)
	}
	if len(out.Corks) != 1 || out.Corks[0].ID != 1 {
		t.Fatalf("want one cork id=1; got %+v", out.Corks)
	}
	if len(out.Heat) != 0 {
		t.Fatalf("near zoom must not return heat; got %d", len(out.Heat))
	}
}

func TestMysteryDelayExcludedFromMapQuery(t *testing.T) {
	now := time.Now()
	bottles := []domain.Bottle{
		{ID: 1, Status: domain.BottleStatusMysteryDelay, IsReleased: false, VisibleAt: now.Add(20 * time.Minute), CurrentLat: 30, CurrentLng: -140},
		{ID: 2, Status: domain.BottleStatusDrifting, IsReleased: true, VisibleAt: now.Add(-time.Hour), CurrentLat: 30.1, CurrentLng: -140.1},
	}
	out := discovery.QueryOcean(6, pacificVP(), bottles)
	if len(out.Corks) != 1 || out.Corks[0].ID != 2 {
		t.Fatalf("Mystery Delay must be invisible; got %+v", out.Corks)
	}
}

func TestSeedsVisibleWithEmptyVisitorOcean(t *testing.T) {
	out := discovery.QueryOcean(6, pacificVP(), discovery.Seeds())
	if len(out.Corks) == 0 {
		t.Fatal("Seed Bottles must appear as Corks in Ocean viewport")
	}
	found := false
	for _, c := range out.Corks {
		if c.IsSeed {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("want at least one is_seed cork; got %+v", out.Corks)
	}
}
