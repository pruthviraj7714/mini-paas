package httpapi

import (
	"mini-paas/internal/handlers"
	"mini-paas/internal/middleware"
	"mini-paas/internal/rabbitmq"
	"mini-paas/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	ProjectService    *service.ProjectService
	DeploymentService *service.DeploymentService
	UserService       *service.UserService
	Producer          *rabbitmq.Producer
}

func NewRouter(deps Dependencies) *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	projectHandler := handlers.NewProjectHandler(deps.ProjectService)
	deploymentHandler := handlers.NewDeploymentHandler(deps.DeploymentService, deps.Producer)
	userHandler := handlers.NewUserHandler(deps.UserService)

	authRouter := router.Group("/auth")
	{
		authRouter.POST("/register", userHandler.Register)
		authRouter.POST("/login", userHandler.Login)
	}

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())

	projectRouter := protected.Group("/projects")
	{
		projectRouter.POST("/", projectHandler.AddProject)
		projectRouter.GET("/", projectHandler.GetProjects)
		projectRouter.GET("/:projectID", projectHandler.GetProject)
		projectRouter.DELETE("/:projectID", projectHandler.DeleteProject)

		projectRouter.POST("/:projectID/deployments", deploymentHandler.CreateDeployment)
		projectRouter.GET("/:projectID/deployments", deploymentHandler.GetDeployments)
	}

	deploymentRouter := protected.Group("/deployments")
	{
		deploymentRouter.GET("/:deploymentID", deploymentHandler.GetDeploymentByID)
		deploymentRouter.PUT("/:deploymentID/status", deploymentHandler.UpdateStatus)
	}

	return router
}
