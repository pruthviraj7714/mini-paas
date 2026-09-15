package worker

import (
	"context"
	"encoding/json"
	"log"
	"mini-paas/internal/events"
	workerhandlers "mini-paas/internal/worker/handlers"
)

type Worker struct {
	deps              Dependencies
	deploymentHandler *workerhandlers.DeploymentHandler
}

func NewWorker(deps Dependencies) *Worker {
	dh := workerhandlers.NewDeploymentHandler(deps.DeploymentRepo, deps.ProjectRepo, deps.DockerManager, deps.WorkspaceManager)
	return &Worker{deps: deps, deploymentHandler: dh}
}

func (w *Worker) Start(ctx context.Context) error {
	msgs, err := w.deps.Consumer.Consume(events.DeploymentCreatedRoutingKey)
	if err != nil {
		return err
	}

	log.Println("worker: waiting for deployment jobs...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			var payload events.DeploymentJobPayload
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				log.Println("worker: bad payload:", err)
				msg.Nack(false, false)
				continue
			}
			if err := w.deploymentHandler.Handle(ctx, payload); err != nil {
				log.Println("worker: deployment job failed:", err)
				msg.Nack(false, true) // requeue on failure
				continue
			}
			msg.Ack(false)
		}
	}
}
