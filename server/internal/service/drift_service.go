package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"

	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/util"
	"go.uber.org/zap"
)

const DriftTickHours = 0.25 // 15 minutes = 6 hours simulated ocean drift


// CurrentZone maps a region of the ocean to a dominant current direction and speed. 
// This is a simplified gyre model, real ocean currents follow these
// Broad circular patterns (gyres) driven by wind and the Coriolis effect, but with lots of local variation.
type currentZone struct {
	minLat, maxLat float64
	minLng, maxLng float64
	bearing        float64 // degrees, 0 = north, 90 = east, etc.
	speedKmH       float64
}

var oceanZones = []currentZone{
	{minLat: 0, maxLat: 60, minLng: -80, maxLng: 0, bearing: 45, speedKmH: 2.5},
	// South Atlantic Gyre (counter-clockwise)
	{minLat: -60, maxLat: 0, minLng: -60, maxLng: 20, bearing: 225, speedKmH: 2.0},
	// North Pacific Gyre (clockwise)
	{minLat: 0, maxLat: 65, minLng: 120, maxLng: -120, bearing: 60, speedKmH: 2.8},
	// South Pacific Gyre (counter-clockwise)
	{minLat: -60, maxLat: 0, minLng: 150, maxLng: -70, bearing: 210, speedKmH: 2.2},
	// Indian Ocean Gyre
	{minLat: -60, maxLat: 25, minLng: 40, maxLng: 120, bearing: 270, speedKmH: 1.8},
	// Default fallback — gentle random drift
	{minLat: -90, maxLat: 90, minLng: -180, maxLng: 180, bearing: 0, speedKmH: 0.5},
}

type DriftService interface {
	Tick(ctx context.Context) error
}

type driftService struct {
	bottles repository.BottleRepository
	events  repository.EventRepository
	log     *zap.Logger
}

func NewDriftService(
	bottles repository.BottleRepository,
	events repository.EventRepository,
	log *zap.Logger,
) DriftService {
	return &driftService{bottles: bottles, events: events, log: log}
}

func (s *driftService) Tick(ctx context.Context) error {
	s.log.Info("drift tick fired")
	// Wire in ListActive once it add that query to sqlc.
	// For now the scheduler runs - you'll see the log every 15 minutes, but it won't actually do anything until ListActive is implemented.
	return nil
}

func (s *driftService) driftOne(ctx context.Context, bottle *domain.Bottle, onDrift func(domain.BottleEvent)) error {
	bearing, speed := dominantCurrent(bottle.CurrentLat, bottle.CurrentLng)

	// Add +=10 degrees to random perturbation to make it less predictable, so path looks organic, not like perfect mathematical circles.
	bearing += rand.Float64()*20 - 10
	bearing = math.Mod(bearing+360, 360)

	newLat, newLng := util.ApplyDrift(bottle.CurrentLat, bottle.CurrentLng, speed, bearing, DriftTickHours)

	event, err := s.events.Create(ctx, repository.CreateEventParams{
		BottleID:  bottle.ID,
		EventType: domain.EventTypeDiscovered,
		Lat:       newLat,
		Lng:       newLng,
	})
	if err != nil {
		return fmt.Errorf("drift bottle %d: %w", bottle.ID, err)
	}

	if onDrift != nil {
		onDrift(*event)
	}

	s.log.Info("bottle drifted", zap.Int32("bottle_id", bottle.ID), zap.Float64("lat", newLat), zap.Float64("lng", newLng))

	return nil
}

func dominantCurrent(lat, lng float64) (bearing, speed float64) {
	for _, z := range oceanZones {
		latIn := lat >= z.minLat && lat <= z.maxLat
		lngIn := false
		if z.minLng <= z.maxLng {
			lngIn = lng >= z.minLng && lng <= z.maxLng
		} else {
			// Zone wraps across the antimeridian (+=180), e.g. North Pacific
			lngIn = lng >= z.minLng || lng <= z.maxLng
		}
		if latIn && lngIn {
			return z.bearing, z.speedKmH
		}
	}
	// Should never happen since the last zone is a global fallback, but just in case:
	last := oceanZones[len(oceanZones)-1]
	return last.bearing, last.speedKmH
}
