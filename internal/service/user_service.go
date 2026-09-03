package service

import (
	"context"
	"errors"
	"mini-paas/internal/models"
	"mini-paas/internal/repository"
	"mini-paas/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, err
		}
	} else {
		return uuid.Nil, errors.New("user already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return uuid.Nil, errors.New("error while hashing password")
	}

	user = &models.User{
		Email:    email,
		Password: hashedPassword,
	}

	userID, err := s.UserRepo.CreateUser(ctx, user)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func (s *UserService) LoginUser(ctx context.Context, email, password string) (string, error) {

	user, err := s.UserRepo.GetUserByEmail(ctx, email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", errors.New("user not found")
	}

	if err != nil {
		return "", err
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("incorrect password")
	}

	token, err := utils.GenerateToken(user.ID.String())
	if err != nil {
		return "", err
	}

	return token, nil
}
