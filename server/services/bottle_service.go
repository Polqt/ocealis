package services

import (
	"context"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/models"
)

type BottleService struct {
	queries *ocealis.Queries
}

func NewBottleService(q *ocealis.Queries) *BottleService {
	return &BottleService{
		queries: q,
	}
}

func toDomainBottle(b ocealis.Bottle) models.Bottle {
	return models.Bottle{
		ID:               int64(b.ID),
		SenderID:         int64(b.SenderID.Int32),
		MessageText:      b.MessageText,
		BottleStyle:      int(b.BottleStyle.Int32),
		StartLat:         b.StartLat.Float64,
		StartLng:         b.StartLng.Float64,
		Hops:             int(b.Hops.Int32),
		ScheduledRelease: b.ScheduledRelease.Time,
		IsReleased:       b.IsRelease.Bool,
		CreatedAt:        b.CreatedAt.Time,
	}
}

func (s *BottleService) GetBottleById(ctx context.Context, id int32) (models.Bottle, error) {
	b, err := s.queries.GetBottle(ctx, id)
	if err != nil {
		return models.Bottle{}, err
	}
	return toDomainBottle(b), nil
}
