package service

import (
	"context"
	"fmt"
	"sort"

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

	// Conver radius to degrees for the bounding box query
	radiusDeg := radius / kmPerDegree

	raw, err := s.bottles.FindNearby(ctx, repository.FindNearbyParams{
		Lat:       input.Lat,
		Lng:       input.Lng,
		RadiusDeg: radiusDeg,
		Cursor:    input.Cursor,
		Limit:     limit * 3, // fetch more than needed, will filter precisely below
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

	// Sort by distance which is the closeest first
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].DistanceKm < filtered[j].DistanceKm
	})

	// Trim to requested limit
	hasMore := len(filtered) > int(limit)
	if hasMore {
		filtered = filtered[:limit]
	}

	result := &domain.CursorResult[BottleWithDistance]{
		Data:    filtered,
		HasMore: hasMore,
	}

	if hasMore && len(filtered) > 0 {
		lastID := filtered[len(filtered)-1].ID
		result.NextCursor = &domain.Cursor{LastID: &lastID}
	}

	return result, nil
}
