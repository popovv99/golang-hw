package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/config"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/metrics"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/storer_config.toml", "Path to storer configuration file")
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

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg.Info("waiting for kafka...")
	if err := kafka.WaitForBroker(ctx, cfg.Kafka.Brokers, kafkaTimeout, cfg.Kafka.Retry); err != nil {
		logg.Error("failed to connect to kafka: " + err.Error())
		os.Exit(1)
	}

	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID)
	if err != nil {
		logg.Error("failed to create kafka consumer: " + err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			logg.Error("failed to close kafka consumer: " + err.Error())
		}
	}()

	logg.Info("calendar storer is running...")

	go func() {
		metricsAddr := os.Getenv("METRICS_ADDR")
		if metricsAddr == "" {
			metricsAddr = ":8082"
		}
		logg.Info("metrics server is starting on " + metricsAddr)
		if err := http.ListenAndServe(metricsAddr, promhttp.Handler()); err != nil {
			logg.Error("metrics server failed: " + err.Error())
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		metrics.Default().IncStorerRuns()

		notification, err := consumer.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logg.Error("failed to read message: " + err.Error())
			metrics.Default().IncStorerErrors()
			continue
		}

		if err := eventStorage.CreateNotification(ctx, notification); err != nil {
			logg.Error("failed to save notification: " + err.Error())
			metrics.Default().IncStorerErrors()
			metrics.Default().IncNotificationsSaved("error")
			continue
		}

		metrics.Default().IncNotificationsSaved("success")
		metrics.Default().SetStorerLastSuccess(time.Now())
		logg.Info("notification saved for event: " + notification.EventID)
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
