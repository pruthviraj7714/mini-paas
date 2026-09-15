package service

import (
	"context"
	"mini-paas/internal/events"
	"mini-paas/internal/models"
	"mini-paas/internal/repository"

	"github.com/google/uuid"
)

type Publisher interface {
	PublishDeploymentJob(ctx context.Context, payload events.DeploymentJobPayload) error
}

type DeploymentService struct {
	deploymentRepo *repository.DeploymentRepository
	projectRepo    *repository.ProjectRepository
	producer       Publisher
}

func NewDeploymentService(dr *repository.DeploymentRepository, pr *repository.ProjectRepository, p Publisher) *DeploymentService {
	return &DeploymentService{deploymentRepo: dr, projectRepo: pr, producer: p}
}

func (s *DeploymentService) CreateDeployment(ctx context.Context, projectID, userID uuid.UUID) (*models.Deployment, string, error) {

	project, err := s.projectRepo.GetProjectByID(ctx, userID, projectID)
	if err != nil {
		return nil, "", err
	}

	deployment := &models.Deployment{
		ProjectID: projectID,
		Status:    models.QUEUED,
	}

	createdDeployment, err := s.deploymentRepo.Create(ctx, deployment)
	if err != nil {
		return nil, "", err
	}

	return createdDeployment, project.RepositoryURL, nil
}

func (s *DeploymentService) GetDeployment(ctx context.Context, userID, deploymentID uuid.UUID) (*models.Deployment, error) {
	return s.deploymentRepo.FindByID(ctx, userID, deploymentID)
}

func (s *DeploymentService) ListProjectDeployments(ctx context.Context, userID, projectID uuid.UUID) ([]*models.Deployment, error) {

	_, err := s.projectRepo.GetProjectByID(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	return s.deploymentRepo.FindByProjectID(ctx, projectID)
}

func (s *DeploymentService) TransitionDeployment() {

}
