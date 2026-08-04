package service_test

import (
	"context"
	"testing"

	"github.com/Polqt/ocealis/internal/discovery"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/service"
)

func TestBrowseMapShowsSeedsWhenOceanEmpty(t *testing.T) {
	svc := service.NewDiscoveryService(&fakeBottles{})
	out, err := svc.BrowseMap(context.Background(), service.BrowseMapInput{
		MinLat: -90, MaxLat: 90, MinLng: -180, MaxLng: 180, Zoom: 6,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Mode != "corks" {
		t.Fatalf("want corks, got %q", out.Mode)
	}
	if len(out.Corks) == 0 {
		t.Fatal("empty visitor Ocean still needs Seed Corks")
	}
	seeded := false
	for _, c := range out.Corks {
		if c.IsSeed {
			seeded = true
			break
		}
	}
	if !seeded {
		t.Fatalf("want Seed Corks; got %+v", out.Corks)
	}
}

func TestBrowseMapRefreshShowsLatestDriftPosition(t *testing.T) {
	repo := &fakeBottles{rows: []domain.Bottle{{
		ID:          42,
		Status:      domain.BottleStatusDrifting,
		IsReleased:  true,
		CurrentLat:  10,
		CurrentLng:  -100,
		BottleStyle: 1,
	}}}
	svc := service.NewDiscoveryService(repo)
	input := service.BrowseMapInput{
		MinLat: 0, MaxLat: 20, MinLng: -110, MaxLng: -90, Zoom: 6,
	}

	first, err := svc.BrowseMap(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	assertCorkPosition(t, first.Corks, 42, 10, -100)

	repo.rows[0].CurrentLat = 10.5
	repo.rows[0].CurrentLng = -99.25
	refreshed, err := svc.BrowseMap(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	assertCorkPosition(t, refreshed.Corks, 42, 10.5, -99.25)
}

func assertCorkPosition(t *testing.T, corks []discovery.Cork, id int32, lat, lng float64) {
	t.Helper()
	for _, cork := range corks {
		if cork.ID != id {
			continue
		}
		if cork.Lat != lat || cork.Lng != lng {
			t.Fatalf("Cork %d at (%f, %f), want (%f, %f)", id, cork.Lat, cork.Lng, lat, lng)
		}
		return
	}
	t.Fatalf("Cork %d missing from map result: %+v", id, corks)
}
