package service

import (
	"context"
	"mini-paas/internal/models"
	"mini-paas/internal/repository"

	"github.com/google/uuid"
)

type DeploymentService struct {
	DeploymentRepo *repository.DeploymentRepository
	ProjectRepo    *repository.ProjectRepository
}

func NewDeploymentService(deploymentRepo *repository.DeploymentRepository, projectRepo *repository.ProjectRepository) *DeploymentService {
	return &DeploymentService{
		DeploymentRepo: deploymentRepo,
		ProjectRepo:    projectRepo,
	}
}

func (s *DeploymentService) CreateDeployment(ctx context.Context, deployment *models.Deployment) (models.DeploymentStatus, error) {

	return s.DeploymentRepo.Create(ctx, deployment)
}

func (s *DeploymentService) GetDeployment(ctx context.Context, userID, deploymentID uuid.UUID) (*models.Deployment, error) {
	return s.DeploymentRepo.FindByID(ctx, userID, deploymentID)
}

func (s *DeploymentService) ListProjectDeployments(ctx context.Context, userID, projectID uuid.UUID) ([]*models.Deployment, error) {
	return s.DeploymentRepo.FindByProjectID(ctx, userID, projectID)
}

func (s *DeploymentService) TransitionDeployment() {

}
