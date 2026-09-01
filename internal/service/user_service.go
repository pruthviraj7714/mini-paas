package service

import (
	"context"
	"mini-paas/internal/repository"

	"github.com/google/uuid"
)

type UserService struct {
	UserRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, email, password string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
