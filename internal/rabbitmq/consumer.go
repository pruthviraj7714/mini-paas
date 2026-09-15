package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	MQ *RabbitMQ
}

func NewConsumer(mq *RabbitMQ) *Consumer {
	return &Consumer{
		MQ: mq,
	}
}

func (c *Consumer) Consume(routingKey string) (<-chan amqp091.Delivery, error) {
	ch := c.MQ.Channel()
	if ch == nil {
		return nil, fmt.Errorf("rabbitmq channel is not available")
	}

	msgs, err := ch.ConsumeWithContext(
		context.Background(),
		routingKey,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// Start consumes DeploymentQueue and reconnects the subscription
// automatically whenever the underlying channel/connection is lost.
func (c *Consumer) Start(ctx context.Context, handler func(context.Context, uuid.UUID) error) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("consumer stopping: context cancelled")
			return ctx.Err()
		default:
		}

		ch := c.MQ.Channel()
		if ch == nil {
			log.Println("rabbitmq channel not available, retrying in 2s...")
			time.Sleep(2 * time.Second)
			continue
		}

		msgs, err := ch.ConsumeWithContext(
			ctx,
			DeploymentQueue,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Printf("failed to register consumer: %v, retrying in 2s...", err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

		// consumeLoop returns when msgs is closed (e.g. connection lost)
		// or ctx is cancelled.
		c.consumeLoop(ctx, msgs, handler)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Println("consumer channel closed, resubscribing in 2s...")
			time.Sleep(2 * time.Second)
		}
	}
}

func (c *Consumer) consumeLoop(ctx context.Context, msgs <-chan amqp091.Delivery, handler func(context.Context, uuid.UUID) error) {
	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return // channel closed, e.g. connection dropped
			}

			executionID, err := uuid.Parse(string(d.Body))
			if err != nil {
				log.Printf("invalid UUID: %v", err)
				continue
			}

			if err := handler(ctx, executionID); err != nil {
				log.Printf("handler error: %v", err)
				continue
			}

			log.Printf(" [x] successfully executed execution with id: %s", d.Body)
		}
	}
}
