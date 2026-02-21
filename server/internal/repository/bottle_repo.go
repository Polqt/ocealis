package repository

import (
	"context"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateBottleParams struct {
	SenderID         int32
	MessageText      string
	BottleStyle      int32
	StartLat         float64
	StartLng         float64
	ScheduledRelease pgtype.Timestamptz
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
}

type postgresBottleRepo struct {
	q *ocealis.Queries
}

func NewBottleRepository(q *ocealis.Queries) BottleRepository {
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
		return nil, err
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
