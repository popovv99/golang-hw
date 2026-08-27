package storage

import "time"

// Notification - сущность уведомления, передаваемая через Kafka и сохраняемая Хранителем.
type Notification struct {
	ID        string    `db:"id"`
	EventID   string    `db:"event_id"`
	Title     string    `db:"title"`
	Date      time.Time `db:"date"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}
