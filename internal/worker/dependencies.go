package worker

import (
	"mini-paas/internal/docker"
	"mini-paas/internal/rabbitmq"
	"mini-paas/internal/repository"
	"mini-paas/internal/workspace"
)

type Dependencies struct {
	Consumer         *rabbitmq.Consumer
	DeploymentRepo   *repository.DeploymentRepository
	ProjectRepo      *repository.ProjectRepository
	DockerManager    *docker.DockerManager
	WorkspaceManager *workspace.WorkspaceManager
}
