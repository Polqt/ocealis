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
	created []repository.CreateEventParams
}

func (r *stampEventsRepo) Create(_ context.Context, params repository.CreateEventParams) (*domain.BottleEvent, error) {
	r.created = append(r.created, params)
	return &domain.BottleEvent{
		ID:        12,
		BottleID:  params.BottleID,
		EventType: params.EventType,
		SealIcon:  params.SealIcon,
		Note:      params.Note,
		CreatedAt: time.Now(),
	}, nil
}
func (r *stampEventsRepo) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return nil, nil
}
func (r *stampEventsRepo) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}
func (r *stampEventsRepo) WithTx(*ocealis.Queries) repository.EventRepository { return r }

func TestStampAppendsJourneyEventAndLeavesBottleDrifting(t *testing.T) {
	bottle := &domain.Bottle{
		ID:         7,
		Status:     domain.BottleStatusDrifting,
		CurrentLat: 30,
		CurrentLng: -140,
	}
	bottles := &openBottleRepo{bottle: bottle}
	events := &stampEventsRepo{}
	svc := service.NewBottleService(nil, bottles, events, nil)

	event, err := svc.StampBottle(context.Background(), service.StampBottleInput{
		BottleID: 7,
		SealIcon: "⚓",
		Note:     "<b>Fair winds</b>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if event.EventType != domain.EventTypeStamp {
		t.Fatalf("want stamp Journey event, got %q", event.EventType)
	}
	if event.SealIcon != "⚓" || event.Note != "Fair winds" {
		t.Fatalf("unexpected Stamp payload: %+v", event)
	}
	if len(events.created) != 1 {
		t.Fatalf("one Stamp action must append one Journey event, got %d", len(events.created))
	}
	if bottles.statusWrites != 0 || bottle.Status != domain.BottleStatusDrifting {
		t.Fatalf("Stamp must leave Bottle drifting; writes=%d status=%q", bottles.statusWrites, bottle.Status)
	}
}
