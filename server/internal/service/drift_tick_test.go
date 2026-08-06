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
	bottles []domain.Bottle
}

func (r *driftBottleRepo) Create(context.Context, repository.CreateBottleParams) (*domain.Bottle, error) {
	return nil, nil
}
func (r *driftBottleRepo) GetByID(_ context.Context, id int32) (*domain.Bottle, error) {
	for i := range r.bottles {
		if r.bottles[i].ID == id {
			return &r.bottles[i], nil
		}
	}
	return nil, nil
}
func (r *driftBottleRepo) UpdateStatus(context.Context, int32, domain.BottleStatus) (*domain.Bottle, error) {
	return nil, nil
}
func (r *driftBottleRepo) UpdatePosition(_ context.Context, id int32, lat, lng float64, status domain.BottleStatus) (*domain.Bottle, error) {
	for i := range r.bottles {
		if r.bottles[i].ID == id {
			r.bottles[i].CurrentLat = lat
			r.bottles[i].CurrentLng = lng
			r.bottles[i].Status = status
			r.bottles[i].Hops++
			return &r.bottles[i], nil
		}
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
	// Deliberately return every life-cycle state. Tick owns the final guard that
	// keeps Mystery Delay Bottles fixed at their Shoreline drop point.
	return append([]domain.Bottle(nil), r.bottles...), nil
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

func TestDriftTickMovesVisibleBottleAndMapReadsPersistedPosition(t *testing.T) {
	bottles := &driftBottleRepo{bottles: []domain.Bottle{
		{
			ID:         11,
			CurrentLat: 30,
			CurrentLng: -60,
			Status:     domain.BottleStatusDrifting,
			IsReleased: true,
		},
		{
			ID:         12,
			CurrentLat: 31,
			CurrentLng: -61,
			Status:     domain.BottleStatusMysteryDelay,
			IsReleased: false,
		},
	}}
	events := &driftEventRepo{}
	drift := service.NewDriftService(
		nil,
		bottles,
		events,
		ws.NewBroadcaster(ws.NewHub(), zap.NewNop()),
		zap.NewNop(),
	)

	if err := drift.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}

	visible := bottles.bottles[0]
	if visible.CurrentLat == 30 && visible.CurrentLng == -60 {
		t.Fatal("scheduled Drift tick did not move the visible Bottle")
	}
	hidden := bottles.bottles[1]
	if hidden.CurrentLat != 31 || hidden.CurrentLng != -61 {
		t.Fatalf("Mystery Delay Bottle must stay at its Shoreline drop point; got (%f, %f)", hidden.CurrentLat, hidden.CurrentLng)
	}
	if len(events.created) != 1 {
		t.Fatalf("want one Drift Journey event, got %d", len(events.created))
	}
	event := events.created[0]
	if event.BottleID != visible.ID || event.EventType != domain.EventTypeDrift {
		t.Fatalf("want Drift Journey event for Bottle %d, got %+v", visible.ID, event)
	}
	if event.Lat != visible.CurrentLat || event.Lng != visible.CurrentLng {
		t.Fatalf("Journey position (%f, %f) differs from persisted map position (%f, %f)",
			event.Lat, event.Lng, visible.CurrentLat, visible.CurrentLng)
	}

	ocean := service.NewDiscoveryService(bottles)
	result, err := ocean.BrowseMap(context.Background(), service.BrowseMapInput{
		MinLat: -90,
		MaxLat: 90,
		MinLng: -180,
		MaxLng: 180,
		Zoom:   6,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, cork := range result.Corks {
		if cork.ID == visible.ID {
			if cork.Lat != visible.CurrentLat || cork.Lng != visible.CurrentLng {
				t.Fatalf("refreshed map has stale Cork position: %+v", cork)
			}
			return
		}
	}
	t.Fatalf("refreshed map is missing visible Cork %d", visible.ID)
}
