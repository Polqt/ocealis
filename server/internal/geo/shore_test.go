package geo_test

import (
	"testing"

	"github.com/Polqt/ocealis/internal/geo"
)

func TestInlandCastSnapsJustOffshore(t *testing.T) {
	// Kansas City — deep inland US
	dropLat, dropLng := geo.ResolveDrop(39.0997, -94.5786)

	if geo.IsLand(dropLat, dropLng) {
		t.Fatalf("drop must be Ocean, got land lat=%v lng=%v", dropLat, dropLng)
	}

	// Must move away from inland origin toward a coast
	if dropLat == 39.0997 && dropLng == -94.5786 {
		t.Fatal("inland cast must not keep inland coordinates")
	}
}

func TestBasinFallbackIsOcean(t *testing.T) {
	p := geo.BasinFallback()
	if geo.IsLand(p.Lat, p.Lng) {
		t.Fatalf("basin fallback must be Ocean, got land %v", p)
	}
}
