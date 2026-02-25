package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateBottleParams struct {
	SenderID         int32
	MessageText      string
	BottleStyle      int32
	StartLat         float64
	StartLng         float64
	IsScheduled      bool
	ScheduledRelease pgtype.Timestamptz
}

type FindNearbyParams struct {
	Lat       float64
	Lng       float64
	RadiusDeg float64 // degree-based bounding box radius for initial filtering; should be larger than the actual desired radius to account for edge cases
	Cursor    *int32
	Limit     int32
}

type BottleRepository interface {
	Create(ctx context.Context, params CreateBottleParams) (*domain.Bottle, error)
	GetByID(ctx context.Context, id int32) (*domain.Bottle, error)
	// UpdateStatus persists a status change (e.g. drifting → discovered).
	UpdateStatus(ctx context.Context, id int32, status domain.BottleStatus) (*domain.Bottle, error)
	// UpdatePosition moves the bottle to new coordinates, increments hops, and sets status.
	UpdatePosition(ctx context.Context, id int32, lat, lng float64, status domain.BottleStatus) (*domain.Bottle, error)
	// ListActive returns all bottles currently drifting that have been released.
	ListActive(ctx context.Context) ([]domain.Bottle, error)
	ReleaseScheduled(ctx context.Context) ([]domain.Bottle, error)
	FindNearby(ctx context.Context, params FindNearbyParams) (*domain.CursorResult[domain.Bottle], error)

	// WithTx returns a new repository instance that uses the provided transaction for all operations.
	WithTx(q *ocealis.Queries) BottleRepository
}

type postgresBottleRepo struct {
	q *ocealis.Queries
}

func NewBottleRepository(q *ocealis.Queries) BottleRepository {
	return &postgresBottleRepo{q: q}
}

func (r *postgresBottleRepo) WithTx(q *ocealis.Queries) BottleRepository {
	return &postgresBottleRepo{q: q}
}

func (r *postgresBottleRepo) Create(ctx context.Context, params CreateBottleParams) (*domain.Bottle, error) {
	row, err := r.q.CreateBottle(ctx, ocealis.CreateBottleParams{
		SenderID:         pgtype.Int4{Int32: params.SenderID, Valid: true},
		MessageText:      params.MessageText,
		BottleStyle:      pgtype.Int4{Int32: params.BottleStyle, Valid: true},
		StartLat:         pgtype.Float8{Float64: params.StartLat, Valid: true},
		StartLng:         pgtype.Float8{Float64: params.StartLng, Valid: true},
		ScheduledRelease: params.ScheduledRelease,
	})
	if err != nil {
		return nil, err
	}
	return mapBottle(row), nil
}

func (r *postgresBottleRepo) GetByID(ctx context.Context, id int32) (*domain.Bottle, error) {
	row, err := r.q.GetBottle(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("bottle not found")
		}
		return nil, fmt.Errorf("get bottle %d:%w", id, err)
	}
	return mapBottle(row), nil
}

func (r *postgresBottleRepo) UpdateStatus(ctx context.Context, id int32, status domain.BottleStatus) (*domain.Bottle, error) {
	row, err := r.q.UpdateBottleStatus(ctx, ocealis.UpdateBottleStatusParams{
		ID:     id,
		Status: string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapBottle(row), nil
}

func (r *postgresBottleRepo) UpdatePosition(ctx context.Context, id int32, lat, lng float64, status domain.BottleStatus) (*domain.Bottle, error) {
	row, err := r.q.UpdateBottlePosition(ctx, ocealis.UpdateBottlePositionParams{
		ID:         id,
		CurrentLat: pgtype.Float8{Float64: lat, Valid: true},
		CurrentLng: pgtype.Float8{Float64: lng, Valid: true},
		Status:     string(status),
	})
	if err != nil {
		return nil, err
	}
	return mapBottle(row), nil
}

func (r *postgresBottleRepo) ListActive(ctx context.Context) ([]domain.Bottle, error) {
	rows, err := r.q.ListActiveDriftingBottles(ctx)
	if err != nil {
		return nil, err
	}
	bottles := make([]domain.Bottle, 0, len(rows))
	for _, row := range rows {
		bottles = append(bottles, *mapBottle(row))
	}
	return bottles, nil
}

func (r *postgresBottleRepo) FindNearby(ctx context.Context, params FindNearbyParams) (*domain.CursorResult[domain.Bottle], error) {
	var cursorID int32
	if params.Cursor != nil {
		cursorID = *params.Cursor
	}

	rows, err := r.q.GetNearbyBottles(ctx, ocealis.GetNearbyBottlesParams{
		Lat:       params.Lat,
		Lng:       params.Lng,
		RadiusDeg: params.RadiusDeg,
		CursorID:  pgtype.Int4{Int32: cursorID, Valid: params.Cursor != nil},
	})
	if err != nil {
		return nil, fmt.Errorf("find nearby bottles: %w", err)
	}

	hasMore := len(rows) > int(params.Limit)
	if hasMore {
		rows = rows[:params.Limit] // trim the extra record
	}

	bottles := make([]domain.Bottle, 0, len(rows))
	for _, row := range rows {
		bottles = append(bottles, *mapBottle(row))
	}

	result := &domain.CursorResult[domain.Bottle]{
		Data:    bottles,
		HasMore: hasMore,
	}

	if hasMore && len(bottles) > 0 {
		lastID := bottles[len(bottles)-1].ID
		result.NextCursor = &domain.Cursor{LastID: &lastID}
	}

	return result, nil
}

func mapBottle(row ocealis.Bottle) *domain.Bottle {
	b := &domain.Bottle{
		ID:          row.ID,
		MessageText: row.MessageText,
		Status:      domain.BottleStatus(row.Status),
	}

	if row.SenderID.Valid {
		b.SenderID = row.SenderID.Int32
	}
	if row.BottleStyle.Valid {
		b.BottleStyle = row.BottleStyle.Int32
	}
	if row.StartLat.Valid {
		b.StartLat = row.StartLat.Float64
	}
	if row.StartLng.Valid {
		b.StartLng = row.StartLng.Float64
	}
	// current position: prefer the tracked column; fall back to start coords.
	if row.CurrentLat.Valid {
		b.CurrentLat = row.CurrentLat.Float64
	} else {
		b.CurrentLat = b.StartLat
	}
	if row.CurrentLng.Valid {
		b.CurrentLng = row.CurrentLng.Float64
	} else {
		b.CurrentLng = b.StartLng
	}
	if row.Hops.Valid {
		b.Hops = row.Hops.Int32
	}
	if row.ScheduledRelease.Valid {
		b.ScheduledRelease = row.ScheduledRelease.Time
	}
	if row.IsRelease.Valid {
		b.IsReleased = row.IsRelease.Bool
	}
	if row.CreatedAt.Valid {
		b.CreatedAt = row.CreatedAt.Time
	}
	return b
}

func (r *postgresBottleRepo) ReleaseScheduled(ctx context.Context) ([]domain.Bottle, error) {
	rows, err := r.q.ListScheduledBottles(ctx)
	if err != nil {
		return nil, err
	}

	bottles := make([]domain.Bottle, 0, len(rows))
	for _, row := range rows {
		bottles = append(bottles, *mapBottle(row))
	}

	return bottles, nil
}
