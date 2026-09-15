package main

import (
	"context"
	"log"
	"mini-paas/internal/config"
	"mini-paas/internal/database"
	"mini-paas/internal/docker"
	"mini-paas/internal/git"
	"mini-paas/internal/rabbitmq"
	"mini-paas/internal/repository"
	"mini-paas/internal/worker"
	"mini-paas/internal/workspace"
	"os"
	"os/signal"
	"syscall"
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
	consumer := rabbitmq.NewConsumer(mq)

	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatal(err)
	}
	defer dockerClient.Close()
	dockerManager := docker.NewManager(dockerClient)

	workspaceManager := workspace.NewWorkspaceManager(&git.Runner{})

	deploymentRepo := repository.NewDeploymentRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	deps := worker.Dependencies{
		Consumer:         consumer,
		DeploymentRepo:   deploymentRepo,
		ProjectRepo:      projectRepo,
		DockerManager:    dockerManager,
		WorkspaceManager: workspaceManager,
	}

	w := worker.NewWorker(deps)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := w.Start(ctx); err != nil {
			log.Fatal("worker failed:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, stopping worker...")
	cancel()
	log.Println("Worker exiting")
}
