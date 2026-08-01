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

type visibilityBottleRepo struct {
	due          domain.Bottle
	visibleCalls int
}

func (r *visibilityBottleRepo) Create(context.Context, repository.CreateBottleParams) (*domain.Bottle, error) {
	return nil, nil
}
func (r *visibilityBottleRepo) GetByID(context.Context, int32) (*domain.Bottle, error) {
	return nil, nil
}
func (r *visibilityBottleRepo) UpdateStatus(context.Context, int32, domain.BottleStatus) (*domain.Bottle, error) {
	return nil, nil
}
func (r *visibilityBottleRepo) UpdatePosition(context.Context, int32, float64, float64, domain.BottleStatus) (*domain.Bottle, error) {
	return nil, nil
}
func (r *visibilityBottleRepo) ReRelease(context.Context, repository.ReReleaseBottleParams) (*domain.Bottle, error) {
	return nil, nil
}
func (r *visibilityBottleRepo) MakeVisible(_ context.Context, id int32) (*domain.Bottle, error) {
	r.visibleCalls++
	r.due.Status = domain.BottleStatusDrifting
	r.due.IsReleased = true
	return &r.due, nil
}
func (r *visibilityBottleRepo) ListActive(context.Context) ([]domain.Bottle, error) {
	return nil, nil
}
func (r *visibilityBottleRepo) ReleaseScheduled(context.Context) ([]domain.Bottle, error) {
	return []domain.Bottle{r.due}, nil
}
func (r *visibilityBottleRepo) FindNearby(context.Context, repository.FindNearbyParams) (*domain.CursorResult[domain.Bottle], error) {
	return nil, nil
}
func (r *visibilityBottleRepo) WithTx(*ocealis.Queries) repository.BottleRepository { return r }

type visibilityEventRepo struct {
	createCalls int
}

func (r *visibilityEventRepo) Create(context.Context, repository.CreateEventParams) (*domain.BottleEvent, error) {
	r.createCalls++
	return nil, nil
}
func (r *visibilityEventRepo) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return nil, nil
}
func (r *visibilityEventRepo) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}
func (r *visibilityEventRepo) WithTx(*ocealis.Queries) repository.EventRepository { return r }

func TestMysteryDelayEndsAtRelocatedPositionWithoutDuplicateJourneyEvent(t *testing.T) {
	bottles := &visibilityBottleRepo{due: domain.Bottle{
		ID:          7,
		StartLat:    30,
		StartLng:    -140,
		CurrentLat:  28.5,
		CurrentLng:  -94.5,
		Status:      domain.BottleStatusMysteryDelay,
		IsReleased:  false,
		MessageText: "unchanged",
	}}
	events := &visibilityEventRepo{}
	svc := service.NewDriftService(
		nil,
		bottles,
		events,
		ws.NewBroadcaster(ws.NewHub(), zap.NewNop()),
		zap.NewNop(),
	)

	if err := svc.ReleaseScheduled(context.Background()); err != nil {
		t.Fatal(err)
	}
	if bottles.visibleCalls != 1 {
		t.Fatalf("want one Mystery Delay visibility flip, got %d", bottles.visibleCalls)
	}
	if bottles.due.CurrentLat != 28.5 || bottles.due.CurrentLng != -94.5 {
		t.Fatalf("Bottle returned to old Cast position: %+v", bottles.due)
	}
	if events.createCalls != 0 {
		t.Fatalf("visibility flip must not append duplicate Cast/Re-release event; got %d", events.createCalls)
	}
}
