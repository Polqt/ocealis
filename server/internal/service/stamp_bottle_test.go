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

type stampEventsRepo struct {
	events []domain.BottleEvent
}

func (r *stampEventsRepo) Create(_ context.Context, params repository.CreateEventParams) (*domain.BottleEvent, error) {
	event := domain.BottleEvent{
		ID:        int32(len(r.events) + 1),
		BottleID:  params.BottleID,
		EventType: params.EventType,
		Lat:       params.Lat,
		Lng:       params.Lng,
		SealIcon:  params.SealIcon,
		Note:      params.Note,
		CreatedAt: time.Now(),
	}
	r.events = append(r.events, event)
	return &event, nil
}

func (r *stampEventsRepo) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return r.events, nil
}

func (r *stampEventsRepo) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}

func (r *stampEventsRepo) WithTx(*ocealis.Queries) repository.EventRepository { return r }

func TestStampAppendsJourneyEventAndLeavesBottleDiscoverable(t *testing.T) {
	bottle := &domain.Bottle{
		ID:          7,
		Nickname:    "shorefox",
		MessageText: "tide took this",
		CurrentLat:  31.2,
		CurrentLng:  -142.4,
		Status:      domain.BottleStatusDrifting,
		IsReleased:  true,
	}
	bottles := &openBottleRepo{bottle: bottle}
	events := &stampEventsRepo{events: []domain.BottleEvent{{
		ID:        1,
		BottleID:  7,
		EventType: domain.EventTypeCast,
		CreatedAt: time.Now().Add(-time.Hour),
	}}}
	svc := service.NewBottleService(nil, bottles, events, nil)

	journey, err := svc.StampBottle(context.Background(), service.StampBottleInput{
		BottleID: 7,
		SealIcon: "anchor",
		Note:     "<b>safe passage</b>",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(journey.Events) != 2 {
		t.Fatalf("want one appended event, got %d total", len(journey.Events))
	}
	got := journey.Events[1]
	if got.EventType != domain.EventTypeStamp || got.SealIcon != "anchor" || got.Note != "safe passage" {
		t.Fatalf("unexpected Stamp event: %+v", got)
	}
	if got.Lat != bottle.CurrentLat || got.Lng != bottle.CurrentLng {
		t.Fatalf("Stamp coordinates changed: got %v,%v", got.Lat, got.Lng)
	}
	if journey.Bottle.Status != domain.BottleStatusDrifting || !journey.Bottle.IsReleased {
		t.Fatalf("Bottle left discovery after Stamp: %+v", journey.Bottle)
	}
	if bottles.statusWrites != 0 || bottles.positionWrites != 0 {
		t.Fatalf("Stamp mutated Bottle: status=%d position=%d", bottles.statusWrites, bottles.positionWrites)
	}
}
