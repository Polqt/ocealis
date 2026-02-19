package repository

import (
	"context"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateBottleParams struct {
	senderID         int32
	MessageText      string
	BottleStyle      int32
	StartLat         float64
	StartLng         float64
	ScheduledRelease pgtype.Timestamptz
}

type BottleRepository interface {
	Create(ctx context.Context, params CreateBottleParams) (*domain.Bottle, error)
	GetByID(ctx context.Context, id int32) (*domain.Bottle, error)
}

type postgresBottleRepo struct {
	q *ocealis.Queries
}

func (r *postgresBottleRepo) Create(ctx context.Context, params CreateBottleParams) (*domain.Bottle, error) {
	row, err := r.q.CreateBottle(ctx, ocealis.CreateBottleParams{
		SenderID:         pgtype.Int4{Int32: params.senderID, Valid: true},
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

func mapBottle(row ocealis.Bottle) *domain.Bottle {
	b := &domain.Bottle{
		ID:          row.ID,
		MessageText: row.MessageText,
		Status:      domain.BottleStatusDrifting,
	}

	if row.SenderID.Valid {
		b.SenderID = row.SenderID.Int32
	}

	if row.BottleStyle.Valid {
		b.BottleStyle = row.BottleStyle.Int32
	}

	if row.StartLat.Valid {
		b.StartLat = row.StartLat.Float64
		b.CurrentLat = row.StartLat.Float64
	}

	if row.StartLng.Valid {
		b.StartLng = row.StartLng.Float64
		b.CurrentLng = row.StartLng.Float64
	}

	if row.Hops.Valid {
		b.Hops = row.Hops.Int32
	}

	if row.ScheduledRelease.Valid {
		b.ScheduledRelease = row.ScheduledRelease.Time
	}

	if row.IsRelease.Valid {
		b.IsReleased = row.IsRelease.Bool
		if b.IsReleased {
			b.Status = domain.BottleStatusReleased
		}
	}

	if row.CreatedAt.Valid {
		b.CreatedAt = row.CreatedAt.Time
	}

	return b
}
