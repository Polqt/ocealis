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
	events []domain.BottleEvent
}

func (r *driftEventRepo) Create(_ context.Context, params repository.CreateEventParams) (*domain.BottleEvent, error) {
	event := domain.BottleEvent{
		ID:        int32(len(r.events) + 1),
		BottleID:  params.BottleID,
		EventType: params.EventType,
		Lat:       params.Lat,
		Lng:       params.Lng,
		CreatedAt: time.Now(),
	}
	r.events = append(r.events, event)
	return &event, nil
}
func (r *driftEventRepo) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return r.events, nil
}
func (r *driftEventRepo) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}
func (r *driftEventRepo) WithTx(*ocealis.Queries) repository.EventRepository { return r }

func TestDriftTickMovesVisibleBottleAndMapReadsPersistedPosition(t *testing.T) {
	visibleStartLat, visibleStartLng := 30.0, -140.0
	hiddenStartLat, hiddenStartLng := 31.0, -141.0
	bottles := &driftBottleRepo{bottles: []domain.Bottle{
		{
			ID: 7, CurrentLat: visibleStartLat, CurrentLng: visibleStartLng,
			Status: domain.BottleStatusDrifting, IsReleased: true,
		},
		{
			ID: 8, CurrentLat: hiddenStartLat, CurrentLng: hiddenStartLng,
			Status: domain.BottleStatusMysteryDelay, IsReleased: false,
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
	if visible.CurrentLat == visibleStartLat && visible.CurrentLng == visibleStartLng {
		t.Fatalf("scheduled Drift tick did not move visible Bottle: %+v", visible)
	}
	hidden := bottles.bottles[1]
	if hidden.CurrentLat != hiddenStartLat || hidden.CurrentLng != hiddenStartLng {
		t.Fatalf("Mystery Delay Bottle must stay at its Shoreline until visible: %+v", hidden)
	}
	if len(events.events) != 1 || events.events[0].BottleID != visible.ID ||
		events.events[0].EventType != domain.EventTypeDrift {
		t.Fatalf("Journey must record only the visible Bottle's Drift: %+v", events.events)
	}
	if events.events[0].Lat != visible.CurrentLat || events.events[0].Lng != visible.CurrentLng {
		t.Fatalf("Journey Drift position differs from Bottle position: %+v vs %+v", events.events[0], visible)
	}

	ocean := service.NewDiscoveryService(bottles)
	result, err := ocean.BrowseMap(context.Background(), service.BrowseMapInput{
		MinLat: 20, MaxLat: 40, MinLng: -160, MaxLng: -120, Zoom: 6,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, cork := range result.Corks {
		if cork.ID == visible.ID {
			if cork.Lat != visible.CurrentLat || cork.Lng != visible.CurrentLng {
				t.Fatalf("map refresh has stale Cork position: %+v want %+v", cork, visible)
			}
			return
		}
	}
	t.Fatalf("map refresh did not return drifted Cork: %+v", result.Corks)
}
