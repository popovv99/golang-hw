package kafka

import (
	"context"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

// WaitForBroker пытается установить TCP-соединение с первым брокером из списка,
// повторяя попытки с бэкоффом. maxRetries=0 означает бесконечные попытки.
func WaitForBroker(ctx context.Context, brokers []string, timeout time.Duration, maxRetries int) error {
	if len(brokers) == 0 {
		return fmt.Errorf("kafka brokers list is empty")
	}

	if maxRetries == 0 {
		maxRetries = -1
	}

	backoff := time.Second
	for attempt := 0; maxRetries < 0 || attempt < maxRetries; attempt++ {
		dialCtx, cancel := context.WithTimeout(ctx, timeout)
		conn, err := kafkago.DialContext(dialCtx, "tcp", brokers[0])
		cancel()
		if err == nil {
			conn.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			if backoff < 30*time.Second {
				backoff *= 2
			}
		}
	}

	return fmt.Errorf("kafka not available after %d retries", maxRetries)
}
