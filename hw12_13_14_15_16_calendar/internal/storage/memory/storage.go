package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Проверка на занятость даты
	for _, e := range s.events {
		if isTimeOverlap(e.Date, e.EndDate, event.Date, event.EndDate) {
			return storage.ErrDateBusy
		}
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; !exists {
		return storage.ErrEventNotFound
	}

	// Проверка на занятость даты (исключая текущее событие)
	for id, e := range s.events {
		if id != event.ID && isTimeOverlap(e.Date, e.EndDate, event.Date, event.EndDate) {
			return storage.ErrDateBusy
		}
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[id]; !exists {
		return storage.ErrEventNotFound
	}

	delete(s.events, id)
	return nil
}

func (s *Storage) ListEventsDay(ctx context.Context, date time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	for _, event := range s.events {
		if (event.Date.Equal(dayStart) || event.Date.After(dayStart)) && event.Date.Before(dayEnd) {
			result = append(result, event)
		}
	}

	return result, nil
}

func (s *Storage) ListEventsWeek(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	weekStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	weekEnd := weekStart.Add(7 * 24 * time.Hour)

	for _, event := range s.events {
		if (event.Date.Equal(weekStart) || event.Date.After(weekStart)) && event.Date.Before(weekEnd) {
			result = append(result, event)
		}
	}

	return result, nil
}

func (s *Storage) ListEventsMonth(ctx context.Context, startDate time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	monthStart := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	for _, event := range s.events {
		if (event.Date.Equal(monthStart) || event.Date.After(monthStart)) && event.Date.Before(monthEnd) {
			result = append(result, event)
		}
	}

	return result, nil
}

func isTimeOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}
