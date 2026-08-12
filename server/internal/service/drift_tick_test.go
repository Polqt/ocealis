package service_test

import (
	"context"
	"testing"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/internal/service"
	"github.com/Polqt/ocealis/ws"
	"go.uber.org/zap"
)

type driftBottleRepo struct {
	active  []domain.Bottle
	updates []domain.Bottle
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
	for _, bottle := range r.active {
		if bottle.ID != id {
			continue
		}
		bottle.CurrentLat = lat
		bottle.CurrentLng = lng
		bottle.Status = status
		bottle.Hops++
		r.updates = append(r.updates, bottle)
		return &bottle, nil
	}
	return nil, nil
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
		BottleID:  params.BottleID,
		EventType: params.EventType,
		Lat:       params.Lat,
		Lng:       params.Lng,
	}, nil
}
func (r *driftEventRepo) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return nil, nil
}
func (r *driftEventRepo) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}
func (r *driftEventRepo) WithTx(*ocealis.Queries) repository.EventRepository { return r }

func TestDriftTickMovesVisibleBottleAndLeavesMysteryDelayFixed(t *testing.T) {
	visible := domain.Bottle{
		ID:         11,
		CurrentLat: 30,
		CurrentLng: -140,
		Status:     domain.BottleStatusDrifting,
		IsReleased: true,
	}
	hidden := domain.Bottle{
		ID:         12,
		CurrentLat: 31,
		CurrentLng: -141,
		Status:     domain.BottleStatusMysteryDelay,
		IsReleased: false,
	}
	bottles := &driftBottleRepo{active: []domain.Bottle{visible, hidden}}
	events := &driftEventRepo{}
	svc := service.NewDriftService(
		nil,
		bottles,
		events,
		ws.NewBroadcaster(ws.NewHub(), zap.NewNop()),
		zap.NewNop(),
	)

	if err := svc.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(bottles.updates) != 1 {
		t.Fatalf("want only the visible Bottle to Drift, got %d updates", len(bottles.updates))
	}
	moved := bottles.updates[0]
	if moved.ID != visible.ID {
		t.Fatalf("Mystery Delay Bottle Drifted: got Bottle %d", moved.ID)
	}
	if moved.CurrentLat == visible.CurrentLat && moved.CurrentLng == visible.CurrentLng {
		t.Fatalf("Drift tick did not move Bottle: %+v", moved)
	}
	if moved.Hops != visible.Hops+1 {
		t.Fatalf("want one Drift hop, got %d", moved.Hops)
	}
	if len(events.created) != 1 {
		t.Fatalf("want one Drift Journey event, got %d", len(events.created))
	}
	event := events.created[0]
	if event.EventType != domain.EventTypeDrift {
		t.Fatalf("want Drift Journey event, got %q", event.EventType)
	}
	if event.BottleID != moved.ID || event.Lat != moved.CurrentLat || event.Lng != moved.CurrentLng {
		t.Fatalf("Journey event does not match persisted Drift: event=%+v Bottle=%+v", event, moved)
	}
}
