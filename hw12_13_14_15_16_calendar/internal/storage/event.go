package storage

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDateBusy      = errors.New("date is busy")
	ErrEventNotFound = errors.New("event not found")
)

type Event struct {
	ID           string        `db:"id"`
	Title        string        `db:"title"`
	Date         time.Time     `db:"date"`
	EndDate      time.Time     `db:"end_date"`
	Description  string        `db:"description"`
	UserID       string        `db:"user_id"`
	NotifyBefore time.Duration `db:"notify_before"` // За сколько времени высылать уведомление
}

type Storage interface {
	CreateEvent(ctx context.Context, event Event) error
	UpdateEvent(ctx context.Context, event Event) error
	DeleteEvent(ctx context.Context, id string) error
	ListEventsDay(ctx context.Context, date time.Time) ([]Event, error)
	ListEventsWeek(ctx context.Context, startDate time.Time) ([]Event, error)
	ListEventsMonth(ctx context.Context, startDate time.Time) ([]Event, error)
}
