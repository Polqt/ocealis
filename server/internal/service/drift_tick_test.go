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

type driftPositionUpdate struct {
	id     int32
	lat    float64
	lng    float64
	status domain.BottleStatus
}

type driftBottleRepo struct {
	active  []domain.Bottle
	updates []driftPositionUpdate
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
	r.updates = append(r.updates, driftPositionUpdate{id: id, lat: lat, lng: lng, status: status})
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
	}, nil
}
func (r *driftEventRepo) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return nil, nil
}
func (r *driftEventRepo) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}
func (r *driftEventRepo) WithTx(*ocealis.Queries) repository.EventRepository { return r }

func TestDriftTickMovesOnlyVisibleBottlesAndRecordsJourneyProgress(t *testing.T) {
	bottles := &driftBottleRepo{active: []domain.Bottle{
		{
			ID:         1,
			CurrentLat: 30,
			CurrentLng: -40,
			Status:     domain.BottleStatusDrifting,
			IsReleased: true,
		},
		{
			ID:         2,
			CurrentLat: 31,
			CurrentLng: -41,
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
		t.Fatalf("want one visible Bottle position update, got %+v", bottles.updates)
	}
	update := bottles.updates[0]
	if update.id != 1 {
		t.Fatalf("Mystery Delay Bottle moved; updated Bottle %d", update.id)
	}
	if update.lat == 30 && update.lng == -40 {
		t.Fatal("visible Bottle did not Drift")
	}
	if update.status != domain.BottleStatusDrifting {
		t.Fatalf("want drifting status, got %q", update.status)
	}

	if len(events.created) != 1 {
		t.Fatalf("want one Drift Journey event, got %+v", events.created)
	}
	event := events.created[0]
	if event.BottleID != update.id || event.EventType != domain.EventTypeDrift {
		t.Fatalf("wrong Drift Journey event: %+v", event)
	}
	if event.Lat != update.lat || event.Lng != update.lng {
		t.Fatalf("Journey position %+v does not match Bottle update %+v", event, update)
	}
}
