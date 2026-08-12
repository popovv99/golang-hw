//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

const defaultAPIURL = "http://localhost:8888"

func apiBaseURL() string {
    return getEnv("CALENDAR_API_URL", defaultAPIURL)
}

func pgDSN() string {
	host := getEnv("PGHOST", "localhost")
	port := getEnv("PGPORT", "5432")
	user := getEnv("PGUSER", "postgres")
	pass := getEnv("PGPASSWORD", "postgres")
	db := getEnv("PGDATABASE", "calendar")
	schema := getEnv("PGSCHEMA", "hwgocalendar")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", user, pass, host, port, db, schema)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type eventRequest struct {
	Title        string  `json:"title"`
	Date         string  `json:"date"`
	EndDate      string  `json:"end_date"`
	Description  string  `json:"description,omitempty"`
	UserID       string  `json:"user_id"`
	NotifyBefore *string `json:"notify_before,omitempty"`
}

type eventResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Date         string  `json:"date"`
	EndDate      string  `json:"end_date"`
	Description  string  `json:"description,omitempty"`
	UserID       string  `json:"user_id"`
	NotifyBefore *string `json:"notify_before,omitempty"`
}

type errorResponse struct {
	Message *string `json:"message,omitempty"`
	Code    *string `json:"code,omitempty"`
}

func TestMain(m *testing.M) {
	if err := waitForAPI(); err != nil {
		fmt.Fprintln(os.Stderr, "API not ready:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func waitForAPI() error {
	client := &http.Client{Timeout: 2 * time.Second}
	checkURL := apiBaseURL() + "/events/day?date=" + url.QueryEscape(time.Now().UTC().Format(time.RFC3339))
	for i := 0; i < 30; i++ {
		resp, err := client.Get(checkURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout")
}

func doRequest(t *testing.T, method, path string, body io.Reader) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, apiBaseURL()+path, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, b
}

func toJSON(v interface{}) io.Reader {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(b)
}

func createEvent(t *testing.T, req eventRequest) eventResponse {
	t.Helper()
	resp, body := doRequest(t, http.MethodPost, "/events", toJSON(req))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create event: expected 201, got %d: %s", resp.StatusCode, string(body))
	}
	var out eventResponse
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal event: %v", err)
	}
	return out
}

func listEvents(t *testing.T, kind, date string) []eventResponse {
	t.Helper()
	path := fmt.Sprintf("/events/%s?date=%s", kind, url.QueryEscape(date))
	resp, body := doRequest(t, http.MethodGet, path, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list %s: expected 200, got %d: %s", kind, resp.StatusCode, string(body))
	}
	var out []eventResponse
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	return out
}

func deleteEvent(t *testing.T, id string) {
	t.Helper()
	resp, _ := doRequest(t, http.MethodDelete, "/events/"+id, nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete event: expected 204, got %d", resp.StatusCode)
	}
}

func countNotifications(ctx context.Context, eventID string) (int, error) {
	conn, err := pgx.Connect(ctx, pgDSN())
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)
	var count int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM notifications WHERE event_id = $1", eventID).Scan(&count)
	return count, err
}

func waitForNotification(ctx context.Context, eventID string) error {
	for i := 0; i < 30; i++ {
		count, err := countNotifications(ctx, eventID)
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("notification not found for event %s", eventID)
}

func ptr(s string) *string { return &s }

func TestCreateAndListEvents(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New().String()
	date := now.Add(1 * time.Hour)
	endDate := now.Add(2 * time.Hour)

	created := createEvent(t, eventRequest{
		Title:       "Integration test event",
		Date:        date.Format(time.RFC3339),
		EndDate:     endDate.Format(time.RFC3339),
		Description: "test",
		UserID:      userID,
	})
	if created.Title != "Integration test event" {
		t.Fatalf("unexpected title: %s", created.Title)
	}
	if created.UserID != userID {
		t.Fatalf("unexpected user id: %s", created.UserID)
	}

	day := listEvents(t, "day", date.Format(time.RFC3339))
	if len(day) != 1 || day[0].ID != created.ID {
		t.Fatalf("day list: expected 1 event with id %s, got %+v", created.ID, day)
	}

	week := listEvents(t, "week", date.Format(time.RFC3339))
	if len(week) != 1 || week[0].ID != created.ID {
		t.Fatalf("week list: expected 1 event, got %+v", week)
	}

	month := listEvents(t, "month", date.Format(time.RFC3339))
	if len(month) != 1 || month[0].ID != created.ID {
		t.Fatalf("month list: expected 1 event, got %+v", month)
	}

	other := listEvents(t, "day", now.Add(24*time.Hour).Format(time.RFC3339))
	if len(other) != 0 {
		t.Fatalf("other day list: expected 0, got %+v", other)
	}

	deleteEvent(t, created.ID)
}

func TestBusinessErrors(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New().String()
	date := now.Add(1 * time.Hour)
	endDate := now.Add(2 * time.Hour)

	created := createEvent(t, eventRequest{
		Title:   "First",
		Date:    date.Format(time.RFC3339),
		EndDate: endDate.Format(time.RFC3339),
		UserID:  userID,
	})

	// overlapping
	resp, body := doRequest(t, http.MethodPost, "/events", toJSON(eventRequest{
		Title:   "Overlapping",
		Date:    date.Format(time.RFC3339),
		EndDate: endDate.Format(time.RFC3339),
		UserID:  userID,
	}))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("overlap: expected 400, got %d", resp.StatusCode)
	}
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if errResp.Code == nil || *errResp.Code != "DATE_BUSY" {
		t.Fatalf("overlap: expected DATE_BUSY, got %+v", errResp)
	}

	// invalid json
	resp, _ = doRequest(t, http.MethodPost, "/events", strings.NewReader("not-json"))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid json: expected 400, got %d", resp.StatusCode)
	}

	// not found
	resp, _ = doRequest(t, http.MethodDelete, "/events/"+uuid.New().String(), nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("not found: expected 404, got %d", resp.StatusCode)
	}

	deleteEvent(t, created.ID)
}

func TestStorerSavesNotification(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	userID := uuid.New().String()
	notify := "5s"
	date := now.Add(10 * time.Second)
	endDate := now.Add(30 * time.Second)

	created := createEvent(t, eventRequest{
		Title:        "Notify me",
		Date:         date.Format(time.RFC3339),
		EndDate:      endDate.Format(time.RFC3339),
		UserID:       userID,
		NotifyBefore: &notify,
	})

	if err := waitForNotification(ctx, created.ID); err != nil {
		t.Fatalf("storer: %v", err)
	}

	deleteEvent(t, created.ID)
}