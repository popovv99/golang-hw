package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/app"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/logger"
	internalhttp "github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/server/http"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configFile)
	if err != nil {
		panic(err)
	}

	logg := logger.New(config.Logger.Level)

	var eventStorage storage.Storage
	switch config.Storage.Type {
	case "memory":
		eventStorage = memorystorage.New()
	case "sql":
		sqlStorage, err := sqlstorage.New(config.Storage.Database)
		if err != nil {
			logg.Error("failed to connect to database: " + err.Error())
			os.Exit(1)
		}
		eventStorage = sqlStorage
		defer func() {
			if err := sqlStorage.Close(context.Background()); err != nil {
				logg.Error("failed to close database: " + err.Error())
			}
		}()
	default:
		logg.Error("unknown storage type: " + config.Storage.Type)
		panic("unknown storage type: " + config.Storage.Type)
	}

	calendar := app.New(logg, eventStorage)

	server := internalhttp.NewServer(logg, calendar)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx, config.Server.Host, config.Server.Port); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
