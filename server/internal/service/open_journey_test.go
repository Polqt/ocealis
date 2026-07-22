package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/internal/service"
)

type openBottleRepo struct {
	bottle       *domain.Bottle
	statusWrites int
}

func (r *openBottleRepo) Create(context.Context, repository.CreateBottleParams) (*domain.Bottle, error) {
	return nil, nil
}
func (r *openBottleRepo) GetByID(context.Context, int32) (*domain.Bottle, error) {
	return r.bottle, nil
}
func (r *openBottleRepo) UpdateStatus(context.Context, int32, domain.BottleStatus) (*domain.Bottle, error) {
	r.statusWrites++
	return r.bottle, nil
}
func (r *openBottleRepo) UpdatePosition(context.Context, int32, float64, float64, domain.BottleStatus) (*domain.Bottle, error) {
	return nil, nil
}
func (r *openBottleRepo) ListActive(context.Context) ([]domain.Bottle, error) { return nil, nil }
func (r *openBottleRepo) ReleaseScheduled(context.Context) ([]domain.Bottle, error) {
	return nil, nil
}
func (r *openBottleRepo) FindNearby(context.Context, repository.FindNearbyParams) (*domain.CursorResult[domain.Bottle], error) {
	return nil, nil
}
func (r *openBottleRepo) WithTx(*ocealis.Queries) repository.BottleRepository { return r }

type journeyEventsRepo struct {
	events []domain.BottleEvent
}

func (r *journeyEventsRepo) Create(context.Context, repository.CreateEventParams) (*domain.BottleEvent, error) {
	return nil, nil
}
func (r *journeyEventsRepo) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return r.events, nil
}
func (r *journeyEventsRepo) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}
func (r *journeyEventsRepo) WithTx(*ocealis.Queries) repository.EventRepository { return r }

func TestOpenDoesNotClaimOrRemoveBottle(t *testing.T) {
	bottle := &domain.Bottle{
		ID:          7,
		Nickname:    "shorefox",
		MessageText: "tide took this",
		Status:      domain.BottleStatusDrifting,
		CreatedAt:   time.Now(),
	}
	bottles := &openBottleRepo{bottle: bottle}
	svc := service.NewBottleService(nil, bottles, &journeyEventsRepo{}, nil)

	got, err := svc.GetBottle(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if got.Nickname != "shorefox" || got.MessageText != "tide took this" {
		t.Fatalf("Open must return Message+Nickname; got %+v", got)
	}
	if got.Status != domain.BottleStatusDrifting {
		t.Fatalf("Open must leave Bottle drifting; got %q", got.Status)
	}
	if bottles.statusWrites != 0 {
		t.Fatalf("Open must not claim/remove; UpdateStatus called %d times", bottles.statusWrites)
	}
}

func TestJourneyEventsAreChronological(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	bottle := &domain.Bottle{
		ID:          3,
		Nickname:    "castaway",
		MessageText: "hello",
		Status:      domain.BottleStatusDrifting,
	}
	// Repo returns newest-first (DESC) — Journey must still read oldest-first.
	events := []domain.BottleEvent{
		{ID: 3, BottleID: 3, EventType: domain.EventTypeStamp, CreatedAt: t0.Add(2 * time.Hour)},
		{ID: 2, BottleID: 3, EventType: domain.EventTypeDrift, CreatedAt: t0.Add(time.Hour)},
		{ID: 1, BottleID: 3, EventType: domain.EventTypeCast, CreatedAt: t0},
	}
	svc := service.NewBottleService(nil, &openBottleRepo{bottle: bottle}, &journeyEventsRepo{events: events}, nil)

	j, err := svc.GetJourney(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if j.Bottle == nil || j.Bottle.MessageText != "hello" {
		t.Fatalf("Journey must include Bottle Message; got %+v", j.Bottle)
	}
	if len(j.Events) != 3 {
		t.Fatalf("want 3 events, got %d", len(j.Events))
	}
	want := []domain.EventType{domain.EventTypeCast, domain.EventTypeDrift, domain.EventTypeStamp}
	for i, typ := range want {
		if j.Events[i].EventType != typ {
			t.Fatalf("Journey order broken at %d: want %q got %q", i, typ, j.Events[i].EventType)
		}
	}
}
