package service

import (
	"context"
	"mini-paas/internal/models"
	"mini-paas/internal/repository"

	"github.com/google/uuid"
)

type ProjectService struct {
	ProjectRepo *repository.ProjectRepository
}

func NewProjectService(projectRepo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		ProjectRepo: projectRepo,
	}
}

func (s *ProjectService) AddProject(ctx context.Context, project *models.Project) (uuid.UUID, error) {
	return s.ProjectRepo.CreateProject(ctx, project)
}

func (s *ProjectService) GetProject(ctx context.Context, userID, projectID uuid.UUID) (*models.Project, error) {
	return s.ProjectRepo.GetProjectByID(ctx, userID, projectID)
}

func (s *ProjectService) GetProjects(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	return s.ProjectRepo.GetProjects(ctx, userID)
}

func (s *ProjectService) DeleteProject(ctx context.Context, userID, projectID uuid.UUID) error {
	return s.ProjectRepo.DeleteProject(ctx, userID, projectID)
}
