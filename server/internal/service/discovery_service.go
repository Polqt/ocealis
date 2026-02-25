package service

import (
	"context"
	"fmt"
	"math"

	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/util"
)

// DiscoveryRadiusKm is the default search radius.
// Bottles within this distance are "discoverable" from a given location.
const DiscoverRadiusKm = 500.0

// kmToDeg converts a kilometer radius to an approximate degree offset.
// Used for the bounding box SQL query — not exact, but fast.
// Exact distance filtering is done in Go after the query.
const kmPerDegree = 111.0

type FindNearbyInput struct {
	Lat      float64
	Lng      float64
	RadiusKm float64 // default to DiscoverRadiusKm if 0
	Cursor   *int32
	Limit    int32
}

type BottleWithDistance struct {
	domain.Bottle
	DistanceKm float64 `json:"distance_km"`
}

type DiscoveryService interface {
	FindNearby(ctx context.Context, input FindNearbyInput) (*domain.CursorResult[BottleWithDistance], error)
}

type discoverService struct {
	bottles repository.BottleRepository
}

func NewDiscoveryService(bottles repository.BottleRepository) DiscoveryService {
	return &discoverService{bottles: bottles}
}

func (s *discoverService) FindNearby(ctx context.Context, input FindNearbyInput) (*domain.CursorResult[BottleWithDistance], error) {
	radius := input.RadiusKm
	if radius == 0 {
		radius = DiscoverRadiusKm
	}

	limit := input.Limit
	if limit == 0 {
		limit = 20
	}

	// Convert radius to a safe degree envelope for both latitude and longitude.
	latDeg := radius / kmPerDegree
	cosLat := math.Cos(input.Lat * math.Pi / 180.0)
	if math.Abs(cosLat) < 1e-6 {
		cosLat = 1e-6 // prevent division by zero at poles, effectively treating longitude as irrelevant there
	}
	lngDeg := radius / (kmPerDegree * math.Abs(cosLat))
	radiusDeg := math.Max(latDeg, lngDeg)

	// Fetch limit+1 from DB to detect hasMore, same pattern as event pagination.
	// We do NOT multiply by 3 here because we're no longer over-filtering.
	// The Haversine pass will discard bounding-box corners but we accept
	// that the page might be slightly shorter than limit as a result.
	// This is correct behavior, the client just gets fewer results, not wrong ones.
	raw, err := s.bottles.FindNearby(ctx, repository.FindNearbyParams{
		Lat:       input.Lat,
		Lng:       input.Lng,
		RadiusDeg: radiusDeg,
		Cursor:    input.Cursor,
		Limit:     int32(limit) + 1, // fetch more than needed, will filter precisely below
	})
	if err != nil {
		return nil, fmt.Errorf("find nearby bottles:%w", err)
	}

	filtered := make([]BottleWithDistance, 0, len(raw.Data))
	for _, b := range raw.Data {
		dist := util.HaversineKm(input.Lat, input.Lng, b.CurrentLat, b.CurrentLng)
		if dist <= radius {
			filtered = append(filtered, BottleWithDistance{
				Bottle:     b,
				DistanceKm: dist,
			})
		}
	}

	hasMore := raw.HasMore

	// Trim to limit in case all rows survived Haversine filtering
	if len(filtered) > int(limit) {
		filtered = filtered[:limit]
		hasMore = true
	}

	result := &domain.CursorResult[BottleWithDistance]{
		Data:    filtered,
		HasMore: hasMore,
	}

	// Cursor comes from raw.NextCursor, the DB page boundary, not the filtered slice.
	// This is the key fix: cursor tracks DB ordering (id DESC), not distance ordering.
	if raw.NextCursor != nil {
		result.NextCursor = raw.NextCursor
	} else if hasMore && len(filtered) > 0 {
		// Fallback: build cursor from last filtered item if raw didn't produce one
		lastID := filtered[len(filtered)-1].ID
		result.NextCursor = &domain.Cursor{LastID: &lastID}
	}
	return result, nil
}
