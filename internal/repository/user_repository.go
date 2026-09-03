package repository

import (
	"context"
	"mini-paas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user *models.User

	err := r.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error

	return user, err
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	if err := r.DB.WithContext(ctx).Create(user).Error; err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}
