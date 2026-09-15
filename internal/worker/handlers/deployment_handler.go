package workerhandlers

import (
	"context"
	"fmt"
	"mini-paas/internal/docker"
	"mini-paas/internal/events"
	"mini-paas/internal/models"
	"mini-paas/internal/repository"
	"mini-paas/internal/workspace"
)

type DeploymentHandler struct {
	deploymentRepo   *repository.DeploymentRepository
	projectRepo      *repository.ProjectRepository
	dockerManager    *docker.DockerManager
	workspaceManager *workspace.WorkspaceManager
}

func NewDeploymentHandler(
	dr *repository.DeploymentRepository,
	pr *repository.ProjectRepository,
	dm *docker.DockerManager,
	wm *workspace.WorkspaceManager,
) *DeploymentHandler {
	return &DeploymentHandler{deploymentRepo: dr, projectRepo: pr, dockerManager: dm, workspaceManager: wm}
}

func (h *DeploymentHandler) Handle(ctx context.Context, payload events.DeploymentJobPayload) error {
	if err := h.deploymentRepo.UpdateStatus(ctx, payload.DeploymentID, models.BUILDING); err != nil {
		return err
	}

	commitSHA := payload.CommitSHA
	if commitSHA == "" {
		commitSHA = "latest"
	}

	workDir, err := h.workspaceManager.PrepareWorkspace(ctx, payload.RepoURL, commitSHA)
	if err != nil {
		h.deploymentRepo.UpdateStatus(ctx, payload.DeploymentID, models.FAILED)
		return err
	}

	imageName := fmt.Sprintf("mini-paas:%s", payload.DeploymentID.String())
	if err := h.dockerManager.Build(ctx, workDir, imageName); err != nil {
		h.deploymentRepo.UpdateStatus(ctx, payload.DeploymentID, models.FAILED)
		return err
	}

	containerName := fmt.Sprintf("container-%s", payload.DeploymentID.String())
	if err := h.dockerManager.Run(ctx, imageName, containerName, ""); err != nil {
		h.deploymentRepo.UpdateStatus(ctx, payload.DeploymentID, models.FAILED)
		return err
	}

	return h.deploymentRepo.UpdateStatus(ctx, payload.DeploymentID, models.RUNNING)
}
