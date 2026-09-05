package main

import (
	"context"
	"log"
	"mini-paas/internal/config"
	"mini-paas/internal/database"
	"mini-paas/internal/handlers"
	"mini-paas/internal/middlewares"
	"mini-paas/internal/repository"
	"mini-paas/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	cfg := config.LoadConfig()

	db, err := database.ConnectPostgres(cfg.DatabaseURL)

	if err != nil {
		panic("error while connecting with database")
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	authRouter := router.Group("/auth")

	{
		authRouter.POST("/register", userHandler.Register)
		authRouter.POST("/login", userHandler.Login)
	}

	projectRepo := repository.NewProjectRepository(db)
	projectService := service.NewProjectService(projectRepo)
	projectHandler := handlers.NewProjectHandler(projectService)

	projectRouter := router.Group("/projects")
	{
		projectRouter.Use(middlewares.AuthMiddleware())
		projectRouter.POST("/", projectHandler.AddProject)
		projectRouter.GET("/", projectHandler.GetProjects)
		projectRouter.GET("/:id", projectHandler.GetProject)
		projectRouter.DELETE("/:id", projectHandler.DeleteProject)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
