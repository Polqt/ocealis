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

type driftBottleRepo struct {
	active  []domain.Bottle
	updates []driftUpdate
}

type driftUpdate struct {
	id     int32
	lat    float64
	lng    float64
	status domain.BottleStatus
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
	r.updates = append(r.updates, driftUpdate{id: id, lat: lat, lng: lng, status: status})
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

func TestDriftTickMovesVisibleBottleAndLeavesMysteryDelayStill(t *testing.T) {
	bottles := &driftBottleRepo{active: []domain.Bottle{
		{
			ID:         7,
			CurrentLat: 30,
			CurrentLng: -140,
			Status:     domain.BottleStatusDrifting,
			IsReleased: true,
		},
		{
			ID:         8,
			CurrentLat: 28.5,
			CurrentLng: -94.5,
			Status:     domain.BottleStatusMysteryDelay,
			IsReleased: false,
		},
	}}
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
		t.Fatalf("want one visible Bottle moved, got updates %+v", bottles.updates)
	}
	moved := bottles.updates[0]
	if moved.id != 7 || moved.status != domain.BottleStatusDrifting {
		t.Fatalf("wrong Bottle drifted: %+v", moved)
	}
	if moved.lat == 30 && moved.lng == -140 {
		t.Fatalf("Drift tick did not move Bottle: %+v", moved)
	}
	if len(events.created) != 1 || events.created[0].EventType != domain.EventTypeDrift {
		t.Fatalf("want one Drift Journey event, got %+v", events.created)
	}
	if events.created[0].Lat != moved.lat || events.created[0].Lng != moved.lng {
		t.Fatalf("Journey Drift position and Bottle position differ: event=%+v update=%+v", events.created[0], moved)
	}
}
