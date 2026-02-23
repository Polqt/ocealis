package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Polqt/ocealis/db"
	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/ws"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pool    *pgxpool.Pool
	bottles repository.BottleRepository
	events  repository.EventRepository
	bc      *ws.Broadcaster
}

func NewBottleService(
	pool *pgxpool.Pool,
	bottles repository.BottleRepository,
	events repository.EventRepository,
	bc *ws.Broadcaster,
) BottleService {
	return &bottleService{pool: pool, bottles: bottles, events: events, bc: bc}
}

func (s *bottleService) CreateBottle(ctx context.Context, input CreateBottleInput) (*domain.Bottle, error) {
	releaseAt := time.Now()
	if input.ReleaseAt != nil {
		releaseAt = *input.ReleaseAt
	}

	var bottle *domain.Bottle

	err := db.WithTransaction(ctx, s.pool, func(q *ocealis.Queries) error {
		bottlesTx := s.bottles.WithTx(q)
		eventsTx := s.events.WithTx(q)

		var err error

		bottle, err := bottlesTx.Create(ctx, repository.CreateBottleParams{
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
			return fmt.Errorf("create bottle:%w", err)
		}

		if _, err = eventsTx.Create(ctx, repository.CreateEventParams{
			BottleID:  bottle.ID,
			EventType: domain.EventTypeReleased,
			Lat:       input.StartLat,
			Lng:       input.StartLng,
		}); err != nil {
			return fmt.Errorf("create release event:%w", err)
		}

		return nil
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
	// Validation: bottle must exist, not already discovered, and discoverer cannot be sender, no mutation risk.
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

	err = db.WithTransaction(ctx, s.pool, func(q *ocealis.Queries) error {
		bottlesTx := s.bottles.WithTx(q)
		eventsTx := s.events.WithTx(q)

		if _, err = eventsTx.Create(ctx, repository.CreateEventParams{
			BottleID:  bottle.ID,
			EventType: domain.EventTypeDiscovered,
			Lat:       input.UserLat,
			Lng:       input.UserLng,
		}); err != nil {
			return fmt.Errorf("create discovered event:%w", err)
		}

		if _, err := bottlesTx.UpdateStatus(ctx, bottle.ID, domain.BottleStatusDiscovered); err != nil {
			return fmt.Errorf("update bottle status:%w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover bottle:%w", err)
	}

	s.bc.BroadcastDiscovered(bottle.ID)
	return s.GetJourney(ctx, input.BottleID)
}

func (s *bottleService) ReleaseBottle(ctx context.Context, bottleID, userID int32, lat, lng float64) (*domain.Bottle, error) {
	bottle, err := s.bottles.GetByID(ctx, bottleID)
	if err != nil {
		return nil, ErrBottleNotFound
	}

	var updated *domain.Bottle

	err = db.WithTransaction(ctx, s.pool, func(q *ocealis.Queries) error {
		bottlesTx := s.bottles.WithTx(q)
		eventsTx := s.events.WithTx(q)

		if _, err := eventsTx.Create(ctx, repository.CreateEventParams{
			BottleID:  bottle.ID,
			EventType: domain.EventTypeReReleased,
			Lat:       lat,
			Lng:       lng,
		}); err != nil {
			return fmt.Errorf("create re-release event:%w", err)
		}

		updated, err = bottlesTx.UpdatePosition(ctx, bottle.ID, lat, lng, domain.BottleStatusDrifting)
		if err != nil {
			return fmt.Errorf("update bottle position:%w", err)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("release bottle:%w", err)
	}

	s.bc.BroadcastReleased(updated.ID)
	return updated, nil
}
