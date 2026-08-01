package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/Polqt/ocealis/db"
	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/cast"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
	"github.com/Polqt/ocealis/internal/rerelease"
	"github.com/Polqt/ocealis/internal/stamp"
	"github.com/Polqt/ocealis/ws"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrBottleNotFound       = errors.New("bottle not found")
	ErrBottleNotAvailable   = errors.New("bottle is not available to re-release")
	ErrAlreadyDiscovered    = errors.New("bottle already discovered")
	ErrSenderCannotDiscover = errors.New("sender cannot discover their own bottle")
)

type CreateBottleInput struct {
	Nickname    string
	MessageText string
	BottleStyle int32
	// StartLat/StartLng nil → BasinFallback inside Cast plan.
	StartLat *float64
	StartLng *float64
}

type DiscoverBottleInput struct {
	BottleID   int32
	DiscoverID int32
	UserLat    float64
	UserLng    float64
}

type StampBottleInput struct {
	BottleID int32
	SealIcon string
	Note     string
}

type ReReleaseBottleInput struct {
	BottleID int32
	Nickname string
	Lat      *float64
	Lng      *float64
}

type BottleService interface {
	CreateBottle(ctx context.Context, input CreateBottleInput) (*domain.Bottle, error)
	GetBottle(ctx context.Context, id int32) (*domain.Bottle, error)
	GetJourney(ctx context.Context, bottleID int32) (*domain.Journey, error)
	StampBottle(ctx context.Context, input StampBottleInput) (*domain.BottleEvent, error)
	DiscoverBottle(ctx context.Context, input DiscoverBottleInput) (*domain.Journey, error)
	ReReleaseBottle(ctx context.Context, input ReReleaseBottleInput) (*domain.Bottle, error)
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
	plan, err := cast.Prepare(input.Nickname, input.MessageText, input.StartLat, input.StartLng, time.Now(), rand.New(rand.NewSource(time.Now().UnixNano())))
	if err != nil {
		return nil, err
	}

	var bottle *domain.Bottle

	err = db.WithTransaction(ctx, s.pool, func(q *ocealis.Queries) error {
		bottlesTx := s.bottles.WithTx(q)
		eventsTx := s.events.WithTx(q)

		var err error
		bottle, err = bottlesTx.Create(ctx, repository.CreateBottleParams{
			SenderID:    0,
			Nickname:    plan.Nickname,
			MessageText: plan.MessageText,
			BottleStyle: input.BottleStyle,
			StartLat:    plan.Lat,
			StartLng:    plan.Lng,
			IsScheduled: true,
			Status:      plan.Status,
			IsReleased:  plan.IsReleased,
			ScheduledRelease: pgtype.Timestamptz{
				Time:  plan.VisibleAt,
				Valid: true,
			},
		})
		if err != nil {
			return fmt.Errorf("create bottle:%w", err)
		}

		// Cast Journey event at drop coords even while invisible (Mystery Delay).
		if _, err = eventsTx.Create(ctx, repository.CreateEventParams{
			BottleID:  bottle.ID,
			EventType: domain.EventTypeCast,
			Lat:       plan.Lat,
			Lng:       plan.Lng,
		}); err != nil {
			return fmt.Errorf("create cast event:%w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// No broadcast during Mystery Delay — Cork appears after scheduler flip.
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

	// Journey reads oldest-first (Cast → … → Sink), even if store returns DESC.
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].ID < events[j].ID
		}
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})

	return &domain.Journey{Bottle: bottle, Events: events}, nil
}

func (s *bottleService) StampBottle(ctx context.Context, input StampBottleInput) (*domain.BottleEvent, error) {
	details, err := stamp.Prepare(input.SealIcon, input.Note)
	if err != nil {
		return nil, err
	}

	bottle, err := s.bottles.GetByID(ctx, input.BottleID)
	if err != nil {
		return nil, ErrBottleNotFound
	}

	event, err := s.events.Create(ctx, repository.CreateEventParams{
		BottleID:  bottle.ID,
		EventType: domain.EventTypeStamp,
		Lat:       bottle.CurrentLat,
		Lng:       bottle.CurrentLng,
		SealIcon:  details.SealIcon,
		Note:      details.Note,
	})
	if err != nil {
		return nil, fmt.Errorf("create stamp event:%w", err)
	}

	return event, nil
}

func (s *bottleService) DiscoverBottle(ctx context.Context, input DiscoverBottleInput) (*domain.Journey, error) {
	// Validation: bottle must exist, not already discovered, and discoverer cannot be sender, no mutation risk.
	bottle, err := s.bottles.GetByID(ctx, input.BottleID)
	if err != nil {
		return nil, ErrBottleNotFound
	}

	if input.DiscoverID != 0 && bottle.SenderID == input.DiscoverID {
		return nil, ErrSenderCannotDiscover
	}

	if bottle.Status == domain.BottleStatusClaimed {
		return nil, ErrAlreadyDiscovered
	}

	err = db.WithTransaction(ctx, s.pool, func(q *ocealis.Queries) error {
		bottlesTx := s.bottles.WithTx(q)
		eventsTx := s.events.WithTx(q)

		if _, err = eventsTx.Create(ctx, repository.CreateEventParams{
			BottleID:  bottle.ID,
			EventType: domain.EventTypeOpenedLegacy,
			Lat:       input.UserLat,
			Lng:       input.UserLng,
		}); err != nil {
			return fmt.Errorf("create discovered event:%w", err)
		}

		if _, err := bottlesTx.UpdateStatus(ctx, bottle.ID, domain.BottleStatusClaimed); err != nil {
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

func (s *bottleService) ReReleaseBottle(ctx context.Context, input ReReleaseBottleInput) (*domain.Bottle, error) {
	bottle, err := s.bottles.GetByID(ctx, input.BottleID)
	if err != nil {
		return nil, ErrBottleNotFound
	}
	if bottle.Status != domain.BottleStatusDrifting || !bottle.IsReleased {
		return nil, ErrBottleNotAvailable
	}

	now := time.Now()
	plan, err := rerelease.Prepare(
		input.Nickname,
		input.Lat,
		input.Lng,
		now,
		rand.New(rand.NewSource(now.UnixNano())),
	)
	if err != nil {
		return nil, err
	}

	var updated *domain.Bottle

	apply := func(q *ocealis.Queries) error {
		bottlesTx := s.bottles.WithTx(q)
		eventsTx := s.events.WithTx(q)

		if _, err := eventsTx.Create(ctx, repository.CreateEventParams{
			BottleID:  bottle.ID,
			EventType: domain.EventTypeReReleased,
			Lat:       plan.Lat,
			Lng:       plan.Lng,
		}); err != nil {
			return fmt.Errorf("create re-release event:%w", err)
		}

		updated, err = bottlesTx.ReRelease(ctx, repository.ReReleaseBottleParams{
			ID:         bottle.ID,
			Nickname:   plan.Nickname,
			Lat:        plan.Lat,
			Lng:        plan.Lng,
			Status:     plan.Status,
			IsReleased: plan.IsReleased,
			VisibleAt:  plan.VisibleAt,
		})
		if err != nil {
			return fmt.Errorf("update bottle position:%w", err)
		}
		return nil
	}

	// Tests use in-memory repositories without a database pool.
	if s.pool == nil {
		err = apply(nil)
	} else {
		err = db.WithTransaction(ctx, s.pool, apply)
	}

	if err != nil {
		return nil, fmt.Errorf("re-release bottle:%w", err)
	}

	// No broadcast during Mystery Delay. The scheduler announces the Cork.
	return updated, nil
}
