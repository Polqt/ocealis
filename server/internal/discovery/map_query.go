package discovery

import (
	"math"

	"github.com/Polqt/ocealis/internal/domain"
)

// CorkZoomMin: zoom at/above this → individual Corks; below → heat density.
const CorkZoomMin = 5.0

// CorkCap limits markers in one viewport response.
const CorkCap = 200

// HeatCellDeg is heat grid size in degrees (~111km).
const HeatCellDeg = 2.0

type Viewport struct {
	MinLat, MaxLat float64
	MinLng, MaxLng float64
}

type Cork struct {
	ID     int32   `json:"id"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	IsSeed bool    `json:"is_seed,omitempty"`
}

type HeatCell struct {
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Count int     `json:"count"`
}

type MapResult struct {
	Mode  string     `json:"mode"` // "heat" | "corks"
	Heat  []HeatCell `json:"heat,omitempty"`
	Corks []Cork     `json:"corks,omitempty"`
}

// QueryOcean builds heat or Corks for a viewport. Mystery Delay / unreleased excluded.
func QueryOcean(zoom float64, vp Viewport, bottles []domain.Bottle) MapResult {
	visible := make([]domain.Bottle, 0, len(bottles))
	for _, b := range bottles {
		if b.Status == domain.BottleStatusMysteryDelay || !b.IsReleased {
			continue
		}
		if !inViewport(b.CurrentLat, b.CurrentLng, vp) {
			continue
		}
		visible = append(visible, b)
	}

	if zoom < CorkZoomMin {
		return MapResult{Mode: "heat", Heat: toHeat(visible)}
	}
	corks := make([]Cork, 0, len(visible))
	for _, b := range visible {
		if len(corks) >= CorkCap {
			break
		}
		corks = append(corks, Cork{
			ID:     b.ID,
			Lat:    b.CurrentLat,
			Lng:    b.CurrentLng,
			IsSeed: b.BottleStyle == SeedStyle,
		})
	}
	return MapResult{Mode: "corks", Corks: corks}
}

func inViewport(lat, lng float64, vp Viewport) bool {
	return lat >= vp.MinLat && lat <= vp.MaxLat && lng >= vp.MinLng && lng <= vp.MaxLng
}

func toHeat(bottles []domain.Bottle) []HeatCell {
	type key struct{ i, j int }
	counts := map[key]int{}
	for _, b := range bottles {
		i := int(math.Floor(b.CurrentLat / HeatCellDeg))
		j := int(math.Floor(b.CurrentLng / HeatCellDeg))
		counts[key{i, j}]++
	}
	out := make([]HeatCell, 0, len(counts))
	for k, n := range counts {
		out = append(out, HeatCell{
			Lat:   (float64(k.i) + 0.5) * HeatCellDeg,
			Lng:   (float64(k.j) + 0.5) * HeatCellDeg,
			Count: n,
		})
	}
	return out
}
