package service

import (
	"context"
	"errors"
	"time"

	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/ws"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrBottleNotFound       = errors.New("bottle not found")
	ErrAlreadyDiscovered    = errors.New("bottle already discovered")
	ErrSenderCannotDiscover = errors.New("sender cannot discover their own bottle")
)

type CreateBottleInput struct {
	SenderID    int32
	MessageText string
	BottleStyle int32
	StartLat    float64
	StartLng    float64
	ReleaseAt   *time.Time
}

type DiscoverBottleInput struct {
	BottleID   int32
	DiscoverID int32
	UserLat    float64
	UserLng    float64
}

type BottleService interface {
	CreateBottle(ctx context.Context, input CreateBottleInput) (*domain.Bottle, error)
	GetBottle(ctx context.Context, id int32) (*domain.Bottle, error)
	GetJourney(ctx context.Context, bottleID int32) (*domain.Journey, error)
	DiscoverBottle(ctx context.Context, input DiscoverBottleInput) (*domain.Journey, error)
	ReleaseBottle(ctx context.Context, bottleID, userID int32, lat, lng float64) (*domain.Bottle, error)
}

type bottleService struct {
	bottles repository.BottleRepository
	events  repository.EventRepository
	bc      *ws.Broadcaster
}

func NewBottleService(
	bottles repository.BottleRepository,
	events repository.EventRepository,
	bc *ws.Broadcaster,
) BottleService {
	return &bottleService{bottles: bottles, events: events, bc: bc}
}

func (s *bottleService) CreateBottle(ctx context.Context, input CreateBottleInput) (*domain.Bottle, error) {
	releaseAt := time.Now()
	if input.ReleaseAt != nil {
		releaseAt = *input.ReleaseAt
	}

	bottle, err := s.bottles.Create(ctx, repository.CreateBottleParams{
		SenderID:    input.SenderID,
		MessageText: input.MessageText,
		BottleStyle: input.BottleStyle,
		StartLat:    input.StartLat,
		StartLng:    input.StartLng,
		ScheduledRelease: pgtype.Timestamptz{
			Time:  releaseAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	// Bug fix: first-creation event is "released", not "re_released".
	_, err = s.events.Create(ctx, repository.CreateEventParams{
		BottleID:  bottle.ID,
		EventType: domain.EventTypeReleased,
		Lat:       input.StartLat,
		Lng:       input.StartLng,
	})
	if err != nil {
		return nil, err
	}

	s.bc.BroadcastReleased(bottle.ID)
	return bottle, nil
}

func (s *bottleService) GetBottle(ctx context.Context, id int32) (*domain.Bottle, error) {
	bottle, err := s.bottles.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBottleNotFound
	}
	return bottle, nil
}

func (s *bottleService) GetJourney(ctx context.Context, bottleID int32) (*domain.Journey, error) {
	bottle, err := s.bottles.GetByID(ctx, bottleID)
	if err != nil {
		return nil, ErrBottleNotFound
	}

	events, err := s.events.GetByBottleID(ctx, bottleID)
	if err != nil {
		return nil, err
	}

	// current_lat/current_lng are now persisted in the DB so the bottle
	// struct already reflects the real position; no event-walk needed.
	return &domain.Journey{Bottle: bottle, Events: events}, nil
}

func (s *bottleService) DiscoverBottle(ctx context.Context, input DiscoverBottleInput) (*domain.Journey, error) {
	bottle, err := s.bottles.GetByID(ctx, input.BottleID)
	if err != nil {
		return nil, ErrBottleNotFound
	}

	if bottle.SenderID == input.DiscoverID {
		return nil, ErrSenderCannotDiscover
	}

	if bottle.Status == domain.BottleStatusDiscovered {
		return nil, ErrAlreadyDiscovered
	}

	_, err = s.events.Create(ctx, repository.CreateEventParams{
		BottleID:  bottle.ID,
		EventType: domain.EventTypeDiscovered,
		Lat:       input.UserLat,
		Lng:       input.UserLng,
	})
	if err != nil {
		return nil, err
	}

	// Persist status change to DB (was only mutating in-memory before).
	if _, err = s.bottles.UpdateStatus(ctx, bottle.ID, domain.BottleStatusDiscovered); err != nil {
		return nil, err
	}

	s.bc.BroadcastDiscovered(bottle.ID)
	return s.GetJourney(ctx, input.BottleID)
}

func (s *bottleService) ReleaseBottle(ctx context.Context, bottleID, userID int32, lat, lng float64) (*domain.Bottle, error) {
	bottle, err := s.bottles.GetByID(ctx, bottleID)
	if err != nil {
		return nil, ErrBottleNotFound
	}

	// Bug fix: event type for re-release is "re_released", not "discovered".
	_, err = s.events.Create(ctx, repository.CreateEventParams{
		BottleID:  bottle.ID,
		EventType: domain.EventTypeReReleased,
		Lat:       lat,
		Lng:       lng,
	})
	if err != nil {
		return nil, err
	}

	// Persist new position and status to DB (was only mutating in-memory before).
	updated, err := s.bottles.UpdatePosition(ctx, bottle.ID, lat, lng, domain.BottleStatusDrifting)
	if err != nil {
		return nil, err
	}

	s.bc.BroadcastReleased(updated.ID)
	return updated, nil
}
