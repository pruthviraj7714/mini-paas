package handlers

import (
	"mini-paas/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DeploymentHandler struct {
	DeploymentService *service.DeploymentService
}

func NewDeploymentHandler(deploymentService *service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{
		DeploymentService: deploymentService,
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

	deployment, err := h.DeploymentService.CreateDeployment(c.Request.Context(), parsedProjectID, parsedID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create deployment",
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
