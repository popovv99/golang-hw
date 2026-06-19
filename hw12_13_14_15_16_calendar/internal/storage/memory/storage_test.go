package memorystorage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/storage"
)

func TestCreateEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:          uuid.New().String(),
		Title:       "Test Event",
		Date:        time.Now().Add(1 * time.Hour),
		EndDate:     time.Now().Add(2 * time.Hour),
		Description: "Test Description",
		UserID:      "user1",
	}

	err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	// Проверка дубликата
	err = s.CreateEvent(ctx, event)
	if err != storage.ErrDateBusy {
		t.Errorf("expected ErrDateBusy, got %v", err)
	}
}

func TestUpdateEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:          uuid.New().String(),
		Title:       "Test Event",
		Date:        time.Now().Add(1 * time.Hour),
		EndDate:     time.Now().Add(2 * time.Hour),
		Description: "Test Description",
		UserID:      "user1",
	}

	err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	event.Title = "Updated Event"
	err = s.UpdateEvent(ctx, event)
	if err != nil {
		t.Fatalf("failed to update event: %v", err)
	}

	// Обновление несуществующего события
	nonExistentEvent := storage.Event{
		ID:      uuid.New().String(),
		Title:   "Non-existent",
		Date:    time.Now().Add(10 * time.Hour),
		EndDate: time.Now().Add(11 * time.Hour),
		UserID:  "user1",
	}
	err = s.UpdateEvent(ctx, nonExistentEvent)
	if err != storage.ErrEventNotFound {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := storage.Event{
		ID:          uuid.New().String(),
		Title:       "Test Event",
		Date:        time.Now().Add(1 * time.Hour),
		EndDate:     time.Now().Add(2 * time.Hour),
		Description: "Test Description",
		UserID:      "user1",
	}

	err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	err = s.DeleteEvent(ctx, event.ID)
	if err != nil {
		t.Fatalf("failed to delete event: %v", err)
	}

	// Удаление несуществующего события
	err = s.DeleteEvent(ctx, event.ID)
	if err != storage.ErrEventNotFound {
		t.Errorf("expected ErrEventNotFound, got %v", err)
	}
}

func TestListEventsDay(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()

	event1 := storage.Event{
		ID:      uuid.New().String(),
		Title:   "Event 1",
		Date:    now.Add(2 * time.Hour),
		EndDate: now.Add(3 * time.Hour),
		UserID:  "user1",
	}

	event2 := storage.Event{
		ID:      uuid.New().String(),
		Title:   "Event 2",
		Date:    now.Add(25 * time.Hour),
		EndDate: now.Add(26 * time.Hour),
		UserID:  "user1",
	}

	if err := s.CreateEvent(ctx, event1); err != nil {
		t.Fatalf("failed to create event1: %v", err)
	}
	if err := s.CreateEvent(ctx, event2); err != nil {
		t.Fatalf("failed to create event2: %v", err)
	}

	events, err := s.ListEventsDay(ctx, now)
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}

func testListEvents(t *testing.T, s *Storage, listFunc func(context.Context, time.Time) ([]storage.Event, error), event2Offset, event2Duration time.Duration, expectedCount int) {
	ctx := context.Background()

	now := time.Now()

	event1 := storage.Event{
		ID:      uuid.New().String(),
		Title:   "Event 1",
		Date:    now.Add(2 * time.Hour),
		EndDate: now.Add(3 * time.Hour),
		UserID:  "user1",
	}

	event2 := storage.Event{
		ID:      uuid.New().String(),
		Title:   "Event 2",
		Date:    now.Add(event2Offset),
		EndDate: now.Add(event2Duration),
		UserID:  "user1",
	}

	if err := s.CreateEvent(ctx, event1); err != nil {
		t.Fatalf("failed to create event1: %v", err)
	}
	if err := s.CreateEvent(ctx, event2); err != nil {
		t.Fatalf("failed to create event2: %v", err)
	}

	events, err := listFunc(ctx, now)
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}

	if len(events) != expectedCount {
		t.Errorf("expected %d events, got %d", expectedCount, len(events))
	}
}

func TestListEventsWeek(t *testing.T) {
	s := New()
	testListEvents(t, s, s.ListEventsWeek, 3*24*time.Hour, 73*time.Hour, 2)
}

func TestListEventsMonth(t *testing.T) {
	s := New()
	testListEvents(t, s, s.ListEventsMonth, 15*24*time.Hour, 361*time.Hour, 2)
}

func TestConcurrentAccess(t *testing.T) {
	s := New()
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 100
	eventsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				// Каждое событие на отдельный день, чтобы избежать пересечений
				event := storage.Event{
					ID:      uuid.New().String(),
					Title:   "Concurrent Event",
					Date:    time.Now().AddDate(0, 0, goroutineID*eventsPerGoroutine+j),
					EndDate: time.Now().AddDate(0, 0, goroutineID*eventsPerGoroutine+j).Add(1 * time.Hour),
					UserID:  "user1",
				}
				if err := s.CreateEvent(ctx, event); err != nil {
					t.Errorf("failed to create event: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Проверяем, что все события созданы (используем ListEventsDay для проверки, что хранилище работает)
	// Поскольку события на разные дни, проверим просто что хранилище не пустое
	events, err := s.ListEventsDay(ctx, time.Now())
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}

	// Проверяем, что хотя бы первое событие создано
	if len(events) < 1 {
		t.Errorf("expected at least 1 event, got %d", len(events))
	}
}
