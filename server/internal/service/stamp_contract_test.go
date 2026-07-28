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

type contractStampEvents struct {
	events []domain.BottleEvent
}

func (r *contractStampEvents) Create(_ context.Context, params repository.CreateEventParams) (*domain.BottleEvent, error) {
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

func (r *contractStampEvents) GetByBottleID(context.Context, int32) ([]domain.BottleEvent, error) {
	return r.events, nil
}

func (r *contractStampEvents) GetPaginated(context.Context, repository.GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	return nil, nil
}

func (r *contractStampEvents) WithTx(*ocealis.Queries) repository.EventRepository {
	return r
}

func TestStampAppendsJourneyWithoutRemovingBottleFromOcean(t *testing.T) {
	bottle := &domain.Bottle{
		ID:         7,
		CurrentLat: 31.2,
		CurrentLng: -142.4,
		Status:     domain.BottleStatusDrifting,
		IsReleased: true,
	}
	bottles := &openBottleRepo{bottle: bottle}
	events := &contractStampEvents{events: []domain.BottleEvent{{
		ID:        1,
		BottleID:  7,
		EventType: domain.EventTypeCast,
		CreatedAt: time.Now().Add(-time.Hour),
	}}}
	svc := service.NewBottleService(nil, bottles, events, nil)

	journey, err := svc.StampBottle(context.Background(), service.StampBottleInput{
		BottleID: 7,
		SealIcon: "anchor",
		Note:     "safe passage",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(journey.Events) != 2 || journey.Events[1].EventType != domain.EventTypeStamp {
		t.Fatalf("Stamp must append one Journey event: %+v", journey.Events)
	}
	if journey.Events[1].SealIcon != "anchor" || journey.Events[1].Note != "safe passage" {
		t.Fatalf("Stamp payload was not preserved: %+v", journey.Events[1])
	}
	if journey.Bottle.Status != domain.BottleStatusDrifting || !journey.Bottle.IsReleased {
		t.Fatalf("Stamp removed Bottle from Ocean: %+v", journey.Bottle)
	}
	if bottles.statusWrites != 0 {
		t.Fatalf("Stamp changed Bottle status %d times", bottles.statusWrites)
	}
}
