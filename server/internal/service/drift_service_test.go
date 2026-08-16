package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/internal/service"
	"github.com/Polqt/ocealis/ws"
	"go.uber.org/zap"
)

type driftPositionWrite struct {
	id     int32
	lat    float64
	lng    float64
	status domain.BottleStatus
}

type driftBottleRepo struct {
	active []domain.Bottle
	writes []driftPositionWrite
}

func (r *driftBottleRepo) Create(context.Context, repository.CreateBottleParams) (*domain.Bottle, error) {
	return nil, nil
}
func (r *driftBottleRepo) GetByID(context.Context, int32) (*domain.Bottle, error) {
	return nil, nil
}
func (r *driftBottleRepo) UpdateStatus(context.Context, int32, domain.BottleStatus) (*domain.Bottle, error) {
	return nil, nil
}
func (r *driftBottleRepo) UpdatePosition(_ context.Context, id int32, lat, lng float64, status domain.BottleStatus) (*domain.Bottle, error) {
	r.writes = append(r.writes, driftPositionWrite{id: id, lat: lat, lng: lng, status: status})
	return &domain.Bottle{ID: id, CurrentLat: lat, CurrentLng: lng, Status: status}, nil
}
func (r *driftBottleRepo) ReRelease(context.Context, repository.ReReleaseBottleParams) (*domain.Bottle, error) {
	return nil, nil
}
func (r *driftBottleRepo) MakeVisible(context.Context, int32) (*domain.Bottle, error) {
	return nil, nil
}
func (r *driftBottleRepo) ListActive(context.Context) ([]domain.Bottle, error) {
	return r.active, nil
}
func (r *driftBottleRepo) ReleaseScheduled(context.Context) ([]domain.Bottle, error) {
	return nil, nil
}
func (r *driftBottleRepo) FindNearby(context.Context, repository.FindNearbyParams) (*domain.CursorResult[domain.Bottle], error) {
	return nil, nil
}
func (r *driftBottleRepo) WithTx(*ocealis.Queries) repository.BottleRepository { return r }

type driftEventRepo struct {
	created []repository.CreateEventParams
}

func (r *driftEventRepo) Create(_ context.Context, params repository.CreateEventParams) (*domain.BottleEvent, error) {
	r.created = append(r.created, params)
	return &domain.BottleEvent{
		ID:        int32(len(r.created)),
		BottleID:  params.BottleID,
		EventType: params.EventType,
		Lat:       params.Lat,
		Lng:       params.Lng,
		CreatedAt: time.Now(),
	}, nil
}
func (r *driftEventRepo) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return nil, nil
}
func (r *driftEventRepo) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}
func (r *driftEventRepo) WithTx(*ocealis.Queries) repository.EventRepository { return r }

func newTestDriftService(bottles repository.BottleRepository, events repository.EventRepository) service.DriftService {
	return service.NewDriftService(
		nil,
		bottles,
		events,
		ws.NewBroadcaster(ws.NewHub(), zap.NewNop()),
		zap.NewNop(),
	)
}

func TestDriftTickMovesVisibleBottleAndAppendsJourneyEvent(t *testing.T) {
	bottle := domain.Bottle{
		ID:          7,
		CurrentLat:  30,
		CurrentLng:  -140,
		Status:      domain.BottleStatusDrifting,
		IsReleased:  true,
		BottleStyle: 2,
	}
	bottles := &driftBottleRepo{active: []domain.Bottle{bottle}}
	events := &driftEventRepo{}

	if err := newTestDriftService(bottles, events).Tick(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(bottles.writes) != 1 {
		t.Fatalf("want one visible Bottle position update, got %d", len(bottles.writes))
	}
	moved := bottles.writes[0]
	if moved.id != bottle.ID || moved.status != domain.BottleStatusDrifting {
		t.Fatalf("unexpected Drift update: %+v", moved)
	}
	if moved.lat == bottle.CurrentLat && moved.lng == bottle.CurrentLng {
		t.Fatalf("Drift tick did not move Bottle from (%f, %f)", bottle.CurrentLat, bottle.CurrentLng)
	}
	if len(events.created) != 1 {
		t.Fatalf("want one Drift Journey event, got %d", len(events.created))
	}
	event := events.created[0]
	if event.EventType != domain.EventTypeDrift || event.Lat != moved.lat || event.Lng != moved.lng {
		t.Fatalf("Drift Journey event does not match position update: event=%+v update=%+v", event, moved)
	}
}

func TestDriftTickLeavesMysteryDelayBottleStillUntilVisible(t *testing.T) {
	hidden := domain.Bottle{
		ID:         8,
		CurrentLat: 28.5,
		CurrentLng: -94.5,
		Status:     domain.BottleStatusMysteryDelay,
		IsReleased: false,
	}
	bottles := &driftBottleRepo{active: []domain.Bottle{hidden}}
	events := &driftEventRepo{}

	if err := newTestDriftService(bottles, events).Tick(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(bottles.writes) != 0 || len(events.created) != 0 {
		t.Fatalf(
			"Bottle must stay still and Journey must stay unchanged during Mystery Delay; writes=%d events=%d",
			len(bottles.writes),
			len(events.created),
		)
	}
}
