package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/config"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/scheduler_config.toml", "Path to scheduler configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		panic(err)
	}

	logg := logger.New(cfg.Logger.Level)

	eventStorage, closeFn, err := newStorage(cfg.Storage, logg)
	if err != nil {
		logg.Error("failed to create storage: " + err.Error())
		os.Exit(1)
	}
	defer closeFn()

	kafkaTimeout, err := time.ParseDuration(cfg.Kafka.Timeout)
	if err != nil {
		logg.Error("invalid kafka timeout: " + err.Error())
		os.Exit(1)
	}

	schedulerInterval, err := time.ParseDuration(cfg.Scheduler.Interval)
	if err != nil {
		logg.Error("invalid scheduler interval: " + err.Error())
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg.Info("waiting for kafka...")
	if err := kafka.WaitForBroker(ctx, cfg.Kafka.Brokers, kafkaTimeout, cfg.Kafka.Retry); err != nil {
		logg.Error("failed to connect to kafka: " + err.Error())
		os.Exit(1)
	}

	producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	if err != nil {
		logg.Error("failed to create kafka producer: " + err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			logg.Error("failed to close kafka producer: " + err.Error())
		}
	}()

	logg.Info("calendar scheduler is running...")

	ticker := time.NewTicker(schedulerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := process(ctx, eventStorage, producer, logg); err != nil {
				logg.Error("scheduler iteration failed: " + err.Error())
			}
		}
	}
}

func newStorage(cfg config.StorageConf, logg logger.Logger) (storage.Storage, func(), error) {
	switch cfg.Type {
	case "memory":
		return memorystorage.New(), func() {}, nil
	case "sql":
		s, err := sqlstorage.New(cfg.Database)
		if err != nil {
			return nil, nil, err
		}
		return s, func() {
			if err := s.Close(context.Background()); err != nil {
				logg.Error("failed to close storage: " + err.Error())
			}
		}, nil
	default:
		return nil, nil, fmt.Errorf("unknown storage type: %s", cfg.Type)
	}
}

func process(ctx context.Context, eventStorage storage.Storage, producer kafka.Producer, logg logger.Logger) error {
	now := time.Now()

	events, err := eventStorage.ListEventsForNotification(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to list events for notification: %w", err)
	}

	for _, event := range events {
		notification := storage.Notification{
			EventID: event.ID,
			Title:   event.Title,
			Date:    event.Date,
			UserID:  event.UserID,
		}

		if err := producer.Send(ctx, notification); err != nil {
			logg.Error("failed to send notification: " + err.Error())
			continue
		}

		logg.Info("notification sent for event: " + event.ID)
	}

	oneYearAgo := now.AddDate(-1, 0, 0)
	if err := eventStorage.DeleteOldEvents(ctx, oneYearAgo); err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	return nil
}
