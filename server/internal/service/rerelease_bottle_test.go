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

type reReleaseBottleRepo struct {
	bottle *domain.Bottle
}

func (r *reReleaseBottleRepo) Create(context.Context, repository.CreateBottleParams) (*domain.Bottle, error) {
	return nil, nil
}
func (r *reReleaseBottleRepo) GetByID(context.Context, int32) (*domain.Bottle, error) {
	return r.bottle, nil
}
func (r *reReleaseBottleRepo) UpdateStatus(context.Context, int32, domain.BottleStatus) (*domain.Bottle, error) {
	return nil, nil
}
func (r *reReleaseBottleRepo) UpdatePosition(context.Context, int32, float64, float64, domain.BottleStatus) (*domain.Bottle, error) {
	return nil, nil
}
func (r *reReleaseBottleRepo) ReRelease(_ context.Context, params repository.ReReleaseBottleParams) (*domain.Bottle, error) {
	r.bottle.Nickname = params.Nickname
	r.bottle.CurrentLat = params.Lat
	r.bottle.CurrentLng = params.Lng
	r.bottle.VisibleAt = params.VisibleAt
	r.bottle.ScheduledRelease = params.VisibleAt
	r.bottle.Status = params.Status
	r.bottle.IsReleased = params.IsReleased
	r.bottle.Hops++
	return r.bottle, nil
}
func (r *reReleaseBottleRepo) MakeVisible(context.Context, int32) (*domain.Bottle, error) {
	return nil, nil
}
func (r *reReleaseBottleRepo) ListActive(context.Context) ([]domain.Bottle, error) {
	return nil, nil
}
func (r *reReleaseBottleRepo) ReleaseScheduled(context.Context) ([]domain.Bottle, error) {
	return nil, nil
}
func (r *reReleaseBottleRepo) FindNearby(context.Context, repository.FindNearbyParams) (*domain.CursorResult[domain.Bottle], error) {
	return nil, nil
}
func (r *reReleaseBottleRepo) WithTx(*ocealis.Queries) repository.BottleRepository { return r }

type reReleaseEventsRepo struct {
	events []domain.BottleEvent
}

func (r *reReleaseEventsRepo) Create(_ context.Context, params repository.CreateEventParams) (*domain.BottleEvent, error) {
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
func (r *reReleaseEventsRepo) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return r.events, nil
}
func (r *reReleaseEventsRepo) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}
func (r *reReleaseEventsRepo) WithTx(*ocealis.Queries) repository.EventRepository { return r }

func TestReReleaseKeepsMessageAppendsJourneyAndHidesOldCork(t *testing.T) {
	originalMessage := "this Message cannot be rewritten"
	bottle := &domain.Bottle{
		ID:          7,
		Nickname:    "caster",
		MessageText: originalMessage,
		CurrentLat:  30,
		CurrentLng:  -140,
		Status:      domain.BottleStatusDrifting,
		IsReleased:  true,
	}
	bottles := &reReleaseBottleRepo{bottle: bottle}
	events := &reReleaseEventsRepo{events: []domain.BottleEvent{
		{ID: 1, BottleID: 7, EventType: domain.EventTypeCast},
		{ID: 2, BottleID: 7, EventType: domain.EventTypeStamp},
	}}
	svc := service.NewBottleService(nil, bottles, events, nil)
	inlandLat, inlandLng := 39.0997, -94.5786
	started := time.Now()

	got, err := svc.ReReleaseBottle(context.Background(), service.ReReleaseBottleInput{
		BottleID: 7,
		Nickname: "finder",
		Lat:      &inlandLat,
		Lng:      &inlandLng,
	})
	if err != nil {
		t.Fatal(err)
	}

	if got.MessageText != originalMessage {
		t.Fatalf("Re-release rewrote original Message: %q", got.MessageText)
	}
	if got.Nickname != "finder" {
		t.Fatalf("Re-release Nickname=%q want finder", got.Nickname)
	}
	if got.Status != domain.BottleStatusMysteryDelay || got.IsReleased {
		t.Fatalf("old Cork must disappear during Mystery Delay: %+v", got)
	}
	if got.VisibleAt.Before(started.Add(15*time.Minute)) || got.VisibleAt.After(time.Now().Add(30*time.Minute)) {
		t.Fatalf("visible_at=%v outside 15–30 minute Mystery Delay", got.VisibleAt)
	}
	if got.CurrentLat == inlandLat && got.CurrentLng == inlandLng {
		t.Fatal("inland finder position was not snapped to Shoreline")
	}
	if len(events.events) != 3 {
		t.Fatalf("want prior Cast+Stamp and appended Re-release, got %+v", events.events)
	}
	if events.events[1].EventType != domain.EventTypeStamp || events.events[2].EventType != domain.EventTypeReReleased {
		t.Fatalf("Journey did not preserve Stamp then append Re-release: %+v", events.events)
	}
	if events.events[2].Lat != got.CurrentLat || events.events[2].Lng != got.CurrentLng {
		t.Fatalf("Journey Re-release coordinates differ from relocated Bottle: %+v vs %+v", events.events[2], got)
	}
}
