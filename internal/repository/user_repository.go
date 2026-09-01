package repository

import (
	"context"
	"mini-paas/internal/models"

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
