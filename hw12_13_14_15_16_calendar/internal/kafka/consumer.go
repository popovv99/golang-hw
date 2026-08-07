package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// Consumer - абстракция подписчика на уведомления из Kafka.
type Consumer interface {
	Read(ctx context.Context) (storage.Notification, error)
	Close() error
}

type consumer struct {
	reader *kafkago.Reader
}

// NewConsumer создаёт Consumer, читающий сообщения из указанного топика в составе группы.
func NewConsumer(brokers []string, topic, groupID string) (Consumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list is empty")
	}
	if topic == "" {
		return nil, fmt.Errorf("kafka topic is empty")
	}
	if groupID == "" {
		return nil, fmt.Errorf("kafka group_id is empty")
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafkago.FirstOffset,
	})

	return &consumer{reader: reader}, nil
}

func (c *consumer) Read(ctx context.Context) (storage.Notification, error) {
	var notification storage.Notification

	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return notification, err
	}

	if err := json.Unmarshal(msg.Value, &notification); err != nil {
		return notification, fmt.Errorf("failed to unmarshal notification: %w", err)
	}

	return notification, nil
}

func (c *consumer) Close() error {
	return c.reader.Close()
}
