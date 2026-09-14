package handlers

import (
	"fmt"
	"log"
	"mini-paas/internal/docker"
	"mini-paas/internal/service"
	"mini-paas/internal/workspace"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DeploymentHandler struct {
	DeploymentService *service.DeploymentService
	WorkspaceManager  *workspace.WorkspaceManager
	DockerManager     *docker.DockerManager
}

func NewDeploymentHandler(deploymentService *service.DeploymentService, workspaceManager *workspace.WorkspaceManager, dockerManager *docker.DockerManager) *DeploymentHandler {
	return &DeploymentHandler{
		DeploymentService: deploymentService,
		WorkspaceManager:  workspaceManager,
		DockerManager:     dockerManager,
	}
}

func (h *DeploymentHandler) CreateDeployment(c *gin.Context) {
	projectID, exists := c.Params.Get("projectID")

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "project id not found",
		})
		return
	}

	parsedProjectID, err := uuid.Parse(projectID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid project id",
		})
		return
	}

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	parsedID, err := uuid.Parse(userID.(string))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
		})
		return
	}

	deployment, repoURL, err := h.DeploymentService.CreateDeployment(c.Request.Context(), parsedProjectID, parsedID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create deployment",
		})
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	projectDir := filepath.Join(
		dir,
		"internal",
		"project",
		deployment.ID.String(),
	)

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create project directory",
		})
		return
	}

	_, err = h.WorkspaceManager.GitRunner.Clone(
		c.Request.Context(),
		repoURL,
		projectDir,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to clone repository",
		})
		return
	}

	imageName := fmt.Sprintf(
		"mini-paas:%s",
		deployment.ID.String(),
	)

	if err := h.DockerManager.Build(
		c.Request.Context(),
		projectDir,
		imageName,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to build image",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     deployment.ID,
		"status": deployment.Status,
	})
}

func (h *DeploymentHandler) GetDeployments(c *gin.Context) {
	projectID, exists := c.Params.Get("projectID")

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "project id not found",
		})
		return
	}

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	parsedID, err := uuid.Parse(userID.(string))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
		})
		return
	}

	parsedProjectID, err := uuid.Parse(projectID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid project id",
		})
		return
	}

	deployments, err := h.DeploymentService.ListProjectDeployments(c.Request.Context(), parsedID, parsedProjectID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"deployments": deployments})
}

func (h *DeploymentHandler) GetDeploymentByID(c *gin.Context) {
	deploymentID, exists := c.Params.Get("deploymentID")

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "deployment id not found",
		})
		return
	}

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	parsedID, err := uuid.Parse(userID.(string))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
		})
		return
	}

	parsedDeploymentID, err := uuid.Parse(deploymentID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid deployment id",
		})
		return
	}

	deployment, err := h.DeploymentService.GetDeployment(c.Request.Context(), parsedID, parsedDeploymentID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"deployment": deployment})
}

func (h *DeploymentHandler) UpdateStatus(c *gin.Context) {

}
