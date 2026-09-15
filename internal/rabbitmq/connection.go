package rabbitmq

import (
	"log"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	mu      sync.RWMutex
	conn    *amqp091.Connection
	channel *amqp091.Channel
	url     string
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	r := &RabbitMQ{url: url}
	if err := r.connect(); err != nil {
		return nil, err
	}
	go r.handleReconnect()
	return r, nil
}

func (r *RabbitMQ) connect() error {
	conn, err := amqp091.Dial(r.url)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	_, err = ch.QueueDeclare(
		DeploymentQueue,
		true, false, false, false, nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	// Declare queue for deployment jobs
	_, err = ch.QueueDeclare(
		"deployment.created",
		true, false, false, false, nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	r.mu.Lock()
	r.conn = conn
	r.channel = ch
	r.mu.Unlock()
	return nil
}

// handleReconnect watches for connection loss and re-dials.
func (r *RabbitMQ) handleReconnect() {
	for {
		r.mu.RLock()
		conn := r.conn
		r.mu.RUnlock()
		if conn == nil {
			return
		}

		closeErr := <-conn.NotifyClose(make(chan *amqp091.Error))
		log.Printf("rabbitmq connection closed: %v — reconnecting", closeErr)

		for {
			if err := r.connect(); err != nil {
				log.Printf("rabbitmq reconnect failed: %v", err)
				continue // consider a backoff/sleep here
			}
			log.Println("rabbitmq reconnected")
			break
		}
	}
}

// Channel returns the current live channel, safe for concurrent use.
func (r *RabbitMQ) Channel() *amqp091.Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channel
}

func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}
