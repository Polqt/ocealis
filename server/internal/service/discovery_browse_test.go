package service_test

import (
	"context"
	"testing"

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
