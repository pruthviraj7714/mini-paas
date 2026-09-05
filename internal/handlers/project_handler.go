package handlers

import (
	"mini-paas/internal/models"
	"mini-paas/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	ProjectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		ProjectService: projectService,
	}
}

func (h *ProjectHandler) AddProject(c *gin.Context) {
	var req models.ProjectAddRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
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

	projectID, err := h.ProjectService.AddProject(c.Request.Context(), &models.Project{
		Name:          req.Name,
		UserID:        parsedID,
		RepositoryURL: req.RepositoryURL,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "project successfully added", "projectID": projectID})
}

func (h *ProjectHandler) GetProject(c *gin.Context) {

	projectID, exists := c.Params.Get("id")

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

	project, err := h.ProjectService.GetProject(c.Request.Context(), parsedID, parsedProjectID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"project": project})
}

func (h *ProjectHandler) GetProjects(c *gin.Context) {
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

	projects, err := h.ProjectService.GetProjects(c.Request.Context(), parsedID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"projects": projects})
}

func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectID, exists := c.Params.Get("id")

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

	err = h.ProjectService.DeleteProject(c.Request.Context(), parsedID, parsedProjectID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "project successfully deleted"})
}
