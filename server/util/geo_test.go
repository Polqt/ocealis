package util

import "testing"

func TestHaversineKmSamePoint(t *testing.T) {
	if d := HaversineKm(10, 20, 10, 20); d != 0 {
		t.Fatalf("expected 0, got %v", d)
	}
}

func TestApplyDriftMovesNorth(t *testing.T) {
	lat, lng := ApplyDrift(0, 0, 100, 0, 1) // 100km north
	if lat <= 0 {
		t.Fatalf("expected latitude to increase, got %v", lat)
	}
	if lng < -1 || lng > 1 {
		t.Fatalf("expected longitude near 0, got %v", lng)
	}
}
