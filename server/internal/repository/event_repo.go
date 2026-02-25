package repository

import (
	"context"
	"fmt"

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

type GetEventParams struct {
	BottleID int32
	Cursor   *int32
	Limit    int32
}

type EventRepository interface {
	Create(ctx context.Context, params CreateEventParams) (*domain.BottleEvent, error)
	GetByBottleID(ctx context.Context, bottleID int32) ([]domain.BottleEvent, error)
	GetPaginated(ctx context.Context, params GetEventParams) (*domain.CursorResult[domain.BottleEvent], error)
	WithTx(q *ocealis.Queries) EventRepository
}

type postgresEventRepo struct {
	q *ocealis.Queries
}

func NewEventRepository(q *ocealis.Queries) EventRepository {
	return &postgresEventRepo{q: q}
}

func (r *postgresEventRepo) WithTx(q *ocealis.Queries) EventRepository {
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

func (r *postgresEventRepo) GetPaginated(ctx context.Context, params GetEventParams) (*domain.CursorResult[domain.BottleEvent], error) {
	var cursorID int32
	if params.Cursor != nil {
		cursorID = *params.Cursor
	}

	rows, err := r.q.GetBottleEventsPaginated(ctx, ocealis.GetBottleEventsPaginatedParams{
		BottleID: pgtype.Int4{Int32: params.BottleID, Valid: true},
		CursorID: pgtype.Int4{Int32: cursorID, Valid: params.Cursor != nil},
	})
	if err != nil {
		return nil, fmt.Errorf("get paginated events: %w", err)
	}

	hasMore := len(rows) > int(params.Limit)
	if hasMore {
		rows = rows[:params.Limit] // Trim to the requested limit
	}

	events := make([]domain.BottleEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, *mapEvent(row))
	}

	result := &domain.CursorResult[domain.BottleEvent]{
		Data:    events,
		HasMore: hasMore,
	}

	// Set the next cursor to the ID of the last event in the current page
	if hasMore && len(events) > 0 {
		lastID := events[len(events)-1].ID
		result.NextCursor = &domain.Cursor{LastID: &lastID}
	}

	return result, nil
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
	if row.CreatedAt.Valid {
		e.CreatedAt = row.CreatedAt.Time
	}

	return e
}
