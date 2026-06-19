package sqlstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/config"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

// pgErrCodeExclusionViolation - код ошибки PostgreSQL при нарушении
// EXCLUSION-ограничения (пересечение интервалов событий).
const pgErrCodeExclusionViolation = "23P01"

const listEventsQuery = `
	SELECT id, title, date, end_date, description, user_id, notify_before
	FROM events
	WHERE date >= $1 AND date < $2
`

func mapExclusionErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgErrCodeExclusionViolation {
		return storage.ErrDateBusy
	}
	return err
}

type Storage struct {
	db *sqlx.DB
}

func New(cfg config.DatabaseConf) (*Storage, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.Schema)

	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Connect(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Storage) Close(ctx context.Context) error {
	return s.db.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Занятость даты гарантируется EXCLUSION-ограничением в БД (атомарно, без гонок).
	query := `
		INSERT INTO events (id, title, date, end_date, description, user_id, notify_before)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx, query,
		event.ID, event.Title, event.Date, event.EndDate,
		event.Description, event.UserID, event.NotifyBefore)
	return mapExclusionErr(err)
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	query := `
		UPDATE events
		SET title = $2, date = $3, end_date = $4, description = $5, user_id = $6, notify_before = $7
		WHERE id = $1
	`

	result, err := s.db.ExecContext(ctx, query,
		event.ID, event.Title, event.Date, event.EndDate,
		event.Description, event.UserID, event.NotifyBefore)
	if err != nil {
		return mapExclusionErr(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return storage.ErrEventNotFound
	}

	return nil
}

func (s *Storage) ListEventsDay(ctx context.Context, date time.Time) ([]storage.Event, error) {

	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	var events []storage.Event
	err := s.db.SelectContext(ctx, &events, listEventsQuery, dayStart, dayEnd)
	return events, err
}

func (s *Storage) ListEventsWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {

	weekStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	weekEnd := weekStart.Add(7 * 24 * time.Hour)

	var events []storage.Event
	err := s.db.SelectContext(ctx, &events, listEventsQuery, weekStart, weekEnd)
	return events, err
}

func (s *Storage) ListEventsMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {

	monthStart := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	var events []storage.Event
	err := s.db.SelectContext(ctx, &events, listEventsQuery, monthStart, monthEnd)
	return events, err
}
