package main

import (
	"context"
	"log"
	"mini-paas/internal/config"
	"mini-paas/internal/database"
	"mini-paas/internal/httpapi"
	"mini-paas/internal/rabbitmq"
	"mini-paas/internal/repository"
	"mini-paas/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("error while connecting with database:", err)
	}

	mq, err := rabbitmq.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := mq.Close(); err != nil {
			log.Println("failed to close RabbitMQ:", err)
		}
	}()
	producer := rabbitmq.NewProducer(mq)

	// repos
	projectRepo := repository.NewProjectRepository(db)
	deploymentRepo := repository.NewDeploymentRepository(db)
	userRepo := repository.NewUserRepository(db)

	// services
	projectService := service.NewProjectService(projectRepo)
	deploymentService := service.NewDeploymentService(deploymentRepo, projectRepo, producer)
	userService := service.NewUserService(userRepo)

	deps := httpapi.Dependencies{
		ProjectService:    projectService,
		DeploymentService: deploymentService,
		UserService:       userService,
		Producer:          producer,
	}

	router := httpapi.NewRouter(deps)

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
