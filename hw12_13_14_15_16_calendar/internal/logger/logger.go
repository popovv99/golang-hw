package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
	Warn(msg string)
	Debug(msg string)
}

type zerologLogger struct {
	logger zerolog.Logger
}

func New(level string) Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var logLevel zerolog.Level
	switch strings.ToLower(level) {
	case "error":
		logLevel = zerolog.ErrorLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "debug":
		logLevel = zerolog.DebugLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
	}

	zlog := log.Output(output)

	return &zerologLogger{
		logger: zlog,
	}
}

func (l *zerologLogger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

func (l *zerologLogger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

func (l *zerologLogger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l *zerologLogger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}
