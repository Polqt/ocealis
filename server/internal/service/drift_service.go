package service

import (
	"context"
	"fmt"
	"math"
	"math/rand"

	"github.com/Polqt/ocealis/db"
	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/util"
	"github.com/Polqt/ocealis/ws"
	"github.com/jackc/pgx/v5/pgxpool"
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
	ReleaseScheduled(ctx context.Context) error
}

type driftService struct {
	pool    *pgxpool.Pool
	bottles repository.BottleRepository
	events  repository.EventRepository
	bc      *ws.Broadcaster
	log     *zap.Logger
}

func NewDriftService(
	pool *pgxpool.Pool,
	bottles repository.BottleRepository,
	events repository.EventRepository,
	bc *ws.Broadcaster,
	log *zap.Logger,
) DriftService {
	return &driftService{pool: pool, bottles: bottles, events: events, bc: bc, log: log}
}

func (s *driftService) Tick(ctx context.Context) error {
	s.log.Info("drift tick fired")

	activeBots, err := s.bottles.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active bottles: %w", err)
	}

	for i := range activeBots {
		b := &activeBots[i]
		if err := s.driftOne(ctx, b, func(e domain.BottleEvent) {
			s.bc.BroadcastDrift(ws.DriftPayload{
				BottleID:    e.BottleID,
				Lat:         e.Lat,
				Lng:         e.Lng,
				Hops:        b.Hops,
				BottleStyle: b.BottleStyle,
				Timestamp:   e.CreatedAt,
			})
		}); err != nil {
			// Log and continue — one bad bottle doesn’t stop the others.
			s.log.Error("driftOne failed", zap.Int32("bottle_id", b.ID), zap.Error(err))
		}
	}
	return nil
}

func (s *driftService) driftOne(ctx context.Context, bottle *domain.Bottle, onDrift func(domain.BottleEvent)) error {
	// Mystery Delay Bottles remain fixed at their Shoreline until they become
	// visible Corks. The repository filters these too; keep the life rule here.
	if bottle.Status != domain.BottleStatusDrifting || !bottle.IsReleased {
		return nil
	}

	bearing, speed := dominantCurrent(bottle.CurrentLat, bottle.CurrentLng)

	// Add ±10 degrees so paths look organic, not like perfect circles.
	bearing += rand.Float64()*20 - 10
	bearing = math.Mod(bearing+360, 360)

	newLat, newLng := util.ApplyDrift(bottle.CurrentLat, bottle.CurrentLng, speed, bearing, DriftTickHours)

	var event *domain.BottleEvent
	var updated *domain.Bottle
	apply := func(q *ocealis.Queries) error {
		bottles := s.bottles
		events := s.events
		if q != nil {
			bottles = bottles.WithTx(q)
			events = events.WithTx(q)
		}

		var err error
		updated, err = bottles.UpdatePosition(ctx, bottle.ID, newLat, newLng, domain.BottleStatusDrifting)
		if err != nil {
			return fmt.Errorf("update Bottle position: %w", err)
		}

		event, err = events.Create(ctx, repository.CreateEventParams{
			BottleID:  bottle.ID,
			EventType: domain.EventTypeDrift,
			Lat:       newLat,
			Lng:       newLng,
		})
		if err != nil {
			return fmt.Errorf("append Drift Journey event: %w", err)
		}
		return nil
	}

	var err error
	if s.pool == nil {
		err = apply(nil)
	} else {
		err = db.WithTransaction(ctx, s.pool, apply)
	}
	if err != nil {
		return fmt.Errorf("drift Bottle %d: %w", bottle.ID, err)
	}

	*bottle = *updated

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

func (s *driftService) ReleaseScheduled(ctx context.Context) error {
	due, err := s.bottles.ReleaseScheduled(ctx)
	if err != nil {
		return fmt.Errorf("list scheduled bottles:%w", err)
	}

	for _, bottle := range due {
		// Cast/Re-release already appended its Journey event and relocated the
		// Bottle. Ending Mystery Delay only makes that current position visible.
		_, err := s.bottles.MakeVisible(ctx, bottle.ID)
		if err != nil {
			s.log.Error("release scheduled bottle failed", zap.Int32("bottle_id", bottle.ID), zap.Error(err))
			continue
		}

		s.bc.BroadcastReleased(bottle.ID)
		s.log.Info("scheduled bottle released", zap.Int32("bottle_id", bottle.ID))
	}

	return nil
}
