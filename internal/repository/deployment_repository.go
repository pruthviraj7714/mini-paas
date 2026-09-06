package repository

import (
	"context"
	"errors"
	"mini-paas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeploymentRepository struct {
	DB *gorm.DB
}

func NewDeploymentRepository(db *gorm.DB) *DeploymentRepository {
	return &DeploymentRepository{
		DB: db,
	}
}

func (r *DeploymentRepository) Create(ctx context.Context, deployment *models.Deployment) (models.DeploymentStatus, error) {
	if err := r.DB.WithContext(ctx).Create(&deployment).Error; err != nil {
		return "", nil
	}

	return deployment.Status, nil
}

func (r *DeploymentRepository) FindByID(ctx context.Context, userID, deploymentID uuid.UUID) (*models.Deployment, error) {
	var deployment *models.Deployment

	if err := r.DB.WithContext(ctx).First(&deployment, deploymentID).Error; err != nil {
		return nil, err
	}

	if err := r.DB.WithContext(ctx).Where("id = ? AND user_id = ?", deploymentID, userID).Error; err != nil {
		return nil, errors.New("deployment not found")
	}

	return deployment, nil
}

func (r *DeploymentRepository) FindByProjectID(ctx context.Context, userID, projectID uuid.UUID) ([]*models.Deployment, error) {
	var deployments []*models.Deployment

	if err := r.DB.WithContext(ctx).Preload("Project").Where("project_id = ? AND user_id = ?", projectID, userID).Find(&deployments).Error; err != nil {
		return nil, err
	}

	return deployments, nil
}
func (r *DeploymentRepository) UpdateStatus() {

}
