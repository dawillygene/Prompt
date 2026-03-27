package queue

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// QueueItem represents a pending change to be synced
type QueueItem struct {
	ID          int64                  `json:"id"`
	Action      string                 `json:"action"`      // create, update, delete
	EntityType  string                 `json:"entity_type"` // prompt, category, tag
	EntityID    int                    `json:"entity_id"`
	Payload     map[string]interface{} `json:"payload"`
	Status      string                 `json:"status"` // pending, synced, conflict
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Queue manages the local sync queue
type Queue struct {
	db *sql.DB
}

// New creates or opens a queue database
func New() (*Queue, error) {
	queueDir, err := queueDir()
	if err != nil {
		return nil, err
	}

	// Ensure queue directory exists
	if err := os.MkdirAll(queueDir, 0700); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(queueDir, "queue.db")

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	queue := &Queue{db: db}

	// Initialize schema if needed
	if err := queue.initSchema(); err != nil {
		return nil, err
	}

	return queue, nil
}

// Close closes the database connection
func (q *Queue) Close() error {
	if q.db != nil {
		return q.db.Close()
	}
	return nil
}

// Add adds a new item to the queue
func (q *Queue) Add(action, entityType string, entityID int, payload map[string]interface{}) (*QueueItem, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	result, err := q.db.Exec(
		`INSERT INTO queue_items (action, entity_type, entity_id, payload, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, 'pending', ?, ?)`,
		action, entityType, entityID, string(payloadJSON), time.Now(), time.Now(),
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &QueueItem{
		ID:        id,
		Action:    action,
		EntityType: entityType,
		EntityID:  entityID,
		Payload:   payload,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// GetPending retrieves all pending items
func (q *Queue) GetPending() ([]QueueItem, error) {
	rows, err := q.db.Query(
		`SELECT id, action, entity_type, entity_id, payload, status, created_at, updated_at
		 FROM queue_items WHERE status = 'pending' ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []QueueItem
	for rows.Next() {
		var qi QueueItem
		var payloadStr string
		if err := rows.Scan(&qi.ID, &qi.Action, &qi.EntityType, &qi.EntityID, &payloadStr, &qi.Status, &qi.CreatedAt, &qi.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(payloadStr), &qi.Payload); err != nil {
			return nil, err
		}

		items = append(items, qi)
	}

	return items, rows.Err()
}

// GetConflicts retrieves all conflicted items
func (q *Queue) GetConflicts() ([]QueueItem, error) {
	rows, err := q.db.Query(
		`SELECT id, action, entity_type, entity_id, payload, status, created_at, updated_at
		 FROM queue_items WHERE status = 'conflict' ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []QueueItem
	for rows.Next() {
		var qi QueueItem
		var payloadStr string
		if err := rows.Scan(&qi.ID, &qi.Action, &qi.EntityType, &qi.EntityID, &payloadStr, &qi.Status, &qi.CreatedAt, &qi.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(payloadStr), &qi.Payload); err != nil {
			return nil, err
		}

		items = append(items, qi)
	}

	return items, rows.Err()
}

// MarkSynced marks an item as synced
func (q *Queue) MarkSynced(id int64) error {
	_, err := q.db.Exec(
		"UPDATE queue_items SET status = 'synced', updated_at = ? WHERE id = ?",
		time.Now(), id,
	)
	return err
}

// MarkConflict marks an item as having a conflict
func (q *Queue) MarkConflict(id int64) error {
	_, err := q.db.Exec(
		"UPDATE queue_items SET status = 'conflict', updated_at = ? WHERE id = ?",
		time.Now(), id,
	)
	return err
}

// Remove removes an item from the queue
func (q *Queue) Remove(id int64) error {
	_, err := q.db.Exec("DELETE FROM queue_items WHERE id = ?", id)
	return err
}

// Count returns the number of pending items
func (q *Queue) Count() (int, error) {
	var count int
	err := q.db.QueryRow("SELECT COUNT(*) FROM queue_items WHERE status = 'pending'").Scan(&count)
	return count, err
}

// ConflictCount returns the number of conflicted items
func (q *Queue) ConflictCount() (int, error) {
	var count int
	err := q.db.QueryRow("SELECT COUNT(*) FROM queue_items WHERE status = 'conflict'").Scan(&count)
	return count, err
}

// Clear removes all items from the queue
func (q *Queue) Clear() error {
	_, err := q.db.Exec("DELETE FROM queue_items")
	return err
}

// initSchema initializes the queue database schema
func (q *Queue) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS queue_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		action TEXT NOT NULL,
		entity_type TEXT NOT NULL,
		entity_id INTEGER NOT NULL,
		payload TEXT,
		status TEXT DEFAULT 'pending',
		created_at DATETIME,
		updated_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_queue_status ON queue_items(status);
	CREATE INDEX IF NOT EXISTS idx_queue_created_at ON queue_items(created_at);
	`

	_, err := q.db.Exec(schema)
	return err
}

// queueDir returns the queue directory path
func queueDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "myprompts", "queue"), nil
}
