CREATE SCHEMA IF NOT EXISTS hwgocalendar;

SET search_path TO hwgocalendar;

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    description TEXT,
    user_id VARCHAR(255) NOT NULL,
    notify_before BIGINT,
    CONSTRAINT events_time_range_check CHECK (end_date > date),
    CONSTRAINT events_no_overlap EXCLUDE USING gist (
        tsrange(date, end_date) WITH &&
    )
);

CREATE INDEX IF NOT EXISTS idx_events_date ON events(date);
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);
