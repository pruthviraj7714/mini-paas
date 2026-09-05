package repository

import (
	"context"
	"mini-paas/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	DB *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{
		DB: db,
	}
}

func (r *ProjectRepository) CreateProject(ctx context.Context, project *models.Project) (uuid.UUID, error) {
	if err := r.DB.WithContext(ctx).Create(&project).Error; err != nil {
		return uuid.Nil, err
	}

	return project.ID, nil
}

func (r *ProjectRepository) GetProjectByID(ctx context.Context, userID, projectID uuid.UUID) (*models.Project, error) {
	var project *models.Project

	if err := r.DB.WithContext(ctx).Where("user_id = ?", userID).First(&project, projectID).Error; err != nil {
		return nil, err
	}

	return project, nil
}

func (r *ProjectRepository) GetProjects(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	var projects []*models.Project

	if err := r.DB.WithContext(ctx).Where("user_id = ?", userID).Find(&projects).Error; err != nil {
		return nil, err
	}

	return projects, nil
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	if err := r.DB.WithContext(ctx).Delete(&models.Project{}, "id = ? AND user_id = ?", projectID, userID).Error; err != nil {
		return err
	}

	return nil
}
