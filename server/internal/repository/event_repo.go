package repository

import (
	"context"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateEventParams struct {
	BottleID  int32
	EventType domain.EventType
	Lat       float64
	Lng       float64
}

type EventRepository interface {
	Create(ctx context.Context, params CreateEventParams) (*domain.BottleEvent, error)
	GetByBottleID(ctx context.Context, bottleID int32) ([]domain.BottleEvent, error)
}

type postgresEventRepo struct {
	q *ocealis.Queries
}

func NewEventRepository(q *ocealis.Queries) EventRepository {
	return &postgresEventRepo{q: q}
}

func (r *postgresEventRepo) Create(ctx context.Context, params CreateEventParams) (*domain.BottleEvent, error) {
	row, err := r.q.CreateBottleEvent(ctx, ocealis.CreateBottleEventParams{
		BottleID:  pgtype.Int4{Int32: params.BottleID, Valid: true},
		EventType: string(params.EventType),
		Lat:       pgtype.Float8{Float64: params.Lat, Valid: true},
		Lng:       pgtype.Float8{Float64: params.Lng, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return mapEvent(row), nil
}

func (r *postgresEventRepo) GetByBottleID(ctx context.Context, bottleID int32) ([]domain.BottleEvent, error) {
	rows, err := r.q.GetBottleEvents(ctx, pgtype.Int4{Int32: bottleID, Valid: true})
	if err != nil {
		return nil, err
	}

	events := make([]domain.BottleEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, *mapEvent(row))
	}

	return events, nil
}

func mapEvent(row ocealis.BottleEvent) *domain.BottleEvent {
	e := &domain.BottleEvent{
		ID:        row.ID,
		EventType: domain.EventType(row.EventType),
	}

	if row.BottleID.Valid {
		e.BottleID = row.BottleID.Int32
	}
	if row.Lat.Valid {
		e.Lat = row.Lat.Float64
	}
	if row.Lng.Valid {
		e.Lng = row.Lng.Float64
	}

	return e
}
