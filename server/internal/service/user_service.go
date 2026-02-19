package service

import (
	"context"

	"github.com/Polqt/ocealis/internal/domain"
	"github.com/Polqt/ocealis/internal/repository"
)

var ErrUserNotFound = "user not found"

type CreateUserInput struct {
	Nickname  string
	AvatarURL string
}

type UserService interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*domain.User, error)
	GetUser(ctx context.Context, id int32) (*domain.User, error)
}

type userService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &userService{
		users: users,
	}
}

func (s *userService) CreateUser(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	user, err := s.users.Create(ctx, input.Nickname, input.AvatarURL)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUser(ctx context.Context, id int32) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
