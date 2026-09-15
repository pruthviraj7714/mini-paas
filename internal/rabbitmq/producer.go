package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"mini-paas/internal/events"

	"github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	MQ *RabbitMQ
}

func NewProducer(mq *RabbitMQ) *Producer {
	return &Producer{
		MQ: mq,
	}
}

func (p *Producer) PublishDeploymentJob(ctx context.Context, payload events.DeploymentJobPayload) error {
	ch := p.MQ.Channel()
	if ch == nil {
		return fmt.Errorf("rabbitmq channel is not available")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx,
		"",                   // Exchange name (empty string uses default direct exchange)
		"deployment.created", // Routing key (queue name)
		false,                // Mandatory
		false,                // Immediate
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish deployment job: %v", err)
		return err
	}

	log.Printf(" [x] Sent deployment job %s", payload.DeploymentID)
	return nil
}
