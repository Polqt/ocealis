package repository

import (
	"context"

	"github.com/Polqt/ocealis/db/ocealis"
	"github.com/Polqt/ocealis/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	Create(ctx context.Context, nickname, avatarURL string) (*domain.User, error)
	GetByID(ctx context.Context, id int32) (*domain.User, error)
}

type postgresUserRepo struct {
	q *ocealis.Queries
}

func NewUserRepository(q *ocealis.Queries) UserRepository {
	return &postgresUserRepo{q: q}
}

func (r *postgresUserRepo) Create(ctx context.Context, nickname, avatarURL string) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, ocealis.CreateUserParams{
		Nickname:  nickname,
		AvatarUrl: pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
	})
	if err != nil {
		return nil, err
	}

	return mapUser(row), nil
}

func (r *postgresUserRepo) GetByID(ctx context.Context, id int32) (*domain.User, error) {
	row, err := r.q.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return mapUser(row), nil
}

func mapUser(row ocealis.User) *domain.User {
	u := &domain.User{
		ID:        row.ID,
		Nickname:  row.Nickname,
		CreatedAt: row.CreatedAt.Time,
	}

	if row.AvatarUrl.Valid {
		u.AvatarURL = row.AvatarUrl.String
	}

	return u
}
