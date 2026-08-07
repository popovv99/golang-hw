package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// Producer - абстракция отправителя уведомлений в Kafka.
type Producer interface {
	Send(ctx context.Context, notification storage.Notification) error
	Close() error
}

type producer struct {
	writer *kafkago.Writer
}

// NewProducer создаёт Producer, работающий с указанными брокерами и топиком.
func NewProducer(brokers []string, topic string) (Producer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list is empty")
	}
	if topic == "" {
		return nil, fmt.Errorf("kafka topic is empty")
	}

	writer := &kafkago.Writer{
		Addr:     kafkago.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafkago.LeastBytes{},
	}

	return &producer{writer: writer}, nil
}

func (p *producer) Send(ctx context.Context, notification storage.Notification) error {
	payload, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	msg := kafkago.Message{
		Value: payload,
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *producer) Close() error {
	return p.writer.Close()
}
