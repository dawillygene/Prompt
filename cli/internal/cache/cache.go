package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Cache manages the local SQLite cache for offline access
type Cache struct {
	db *sql.DB
}

// Prompt represents a cached prompt
type Prompt struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	Summary      string `json:"summary"`
	Content      string `json:"content"`
	CategoryID   *int   `json:"category_id"`
	Visibility   string `json:"visibility"`
	IsFavorite   bool   `json:"is_favorite"`
	IsArchived   bool   `json:"is_archived"`
	UsageCount   int    `json:"usage_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Category represents a cached category
type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Tag represents a cached tag
type Tag struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// New creates or opens a cache database
func New() (*Cache, error) {
	cacheDir, err := cacheDir()
	if err != nil {
		return nil, err
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(cacheDir, "cache.db")

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	cache := &Cache{db: db}

	// Initialize schema if needed
	if err := cache.initSchema(); err != nil {
		return nil, err
	}

	return cache, nil
}

// Close closes the database connection
func (c *Cache) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// GetPrompts retrieves all cached prompts
func (c *Cache) GetPrompts() ([]Prompt, error) {
	rows, err := c.db.Query("SELECT id, title, slug, summary, content, category_id, visibility, is_favorite, is_archived, usage_count, created_at, updated_at FROM prompts ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prompts []Prompt
	for rows.Next() {
		var p Prompt
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.Content, &p.CategoryID, &p.Visibility, &p.IsFavorite, &p.IsArchived, &p.UsageCount, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}

	return prompts, rows.Err()
}

// GetPrompt retrieves a single cached prompt by ID or slug
func (c *Cache) GetPrompt(ref string) (*Prompt, error) {
	p := &Prompt{}
	err := c.db.QueryRow(
		"SELECT id, title, slug, summary, content, category_id, visibility, is_favorite, is_archived, usage_count, created_at, updated_at FROM prompts WHERE id = ? OR slug = ? LIMIT 1",
		ref, ref,
	).Scan(&p.ID, &p.Title, &p.Slug, &p.Summary, &p.Content, &p.CategoryID, &p.Visibility, &p.IsFavorite, &p.IsArchived, &p.UsageCount, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("prompt not found in cache")
	}
	return p, err
}

// SavePrompt saves or updates a prompt in the cache
func (c *Cache) SavePrompt(p Prompt) error {
	_, err := c.db.Exec(
		`INSERT OR REPLACE INTO prompts (id, title, slug, summary, content, category_id, visibility, is_favorite, is_archived, usage_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Title, p.Slug, p.Summary, p.Content, p.CategoryID, p.Visibility, p.IsFavorite, p.IsArchived, p.UsageCount, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

// DeletePrompt removes a prompt from the cache
func (c *Cache) DeletePrompt(id int) error {
	_, err := c.db.Exec("DELETE FROM prompts WHERE id = ?", id)
	return err
}

// GetCategories retrieves all cached categories
func (c *Cache) GetCategories() ([]Category, error) {
	rows, err := c.db.Query("SELECT id, name, slug, description, created_at, updated_at FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

// SaveCategory saves or updates a category in the cache
func (c *Cache) SaveCategory(cat Category) error {
	_, err := c.db.Exec(
		`INSERT OR REPLACE INTO categories (id, name, slug, description, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		cat.ID, cat.Name, cat.Slug, cat.Description, cat.CreatedAt, cat.UpdatedAt,
	)
	return err
}

// GetTags retrieves all cached tags
func (c *Cache) GetTags() ([]Tag, error) {
	rows, err := c.db.Query("SELECT id, name, slug, description, created_at, updated_at FROM tags ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Description, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}

	return tags, rows.Err()
}

// SaveTag saves or updates a tag in the cache
func (c *Cache) SaveTag(t Tag) error {
	_, err := c.db.Exec(
		`INSERT OR REPLACE INTO tags (id, name, slug, description, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.ID, t.Name, t.Slug, t.Description, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

// GetLastSyncTime retrieves the last sync timestamp
func (c *Cache) GetLastSyncTime() (time.Time, error) {
	var syncTime sql.NullTime
	err := c.db.QueryRow("SELECT value FROM sync_metadata WHERE key = 'last_sync'").Scan(&syncTime)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return syncTime.Time, nil
}

// SetLastSyncTime updates the last sync timestamp
func (c *Cache) SetLastSyncTime(t time.Time) error {
	_, err := c.db.Exec(
		`INSERT OR REPLACE INTO sync_metadata (key, value) VALUES ('last_sync', ?)`,
		t,
	)
	return err
}

// initSchema initializes the cache database schema
func (c *Cache) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS prompts (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		slug TEXT NOT NULL,
		summary TEXT,
		content TEXT NOT NULL,
		category_id INTEGER,
		visibility TEXT DEFAULT 'private',
		is_favorite BOOLEAN DEFAULT 0,
		is_archived BOOLEAN DEFAULT 0,
		usage_count INTEGER DEFAULT 0,
		created_at TEXT,
		updated_at TEXT
	);

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		description TEXT,
		created_at TEXT,
		updated_at TEXT
	);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		description TEXT,
		created_at TEXT,
		updated_at TEXT
	);

	CREATE TABLE IF NOT EXISTS sync_metadata (
		key TEXT PRIMARY KEY,
		value TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_prompts_slug ON prompts(slug);
	CREATE INDEX IF NOT EXISTS idx_prompts_category_id ON prompts(category_id);
	CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);
	CREATE INDEX IF NOT EXISTS idx_tags_slug ON tags(slug);
	`

	_, err := c.db.Exec(schema)
	return err
}

// cacheDir returns the cache directory path
func cacheDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "myprompts", "cache"), nil
}

// ExportCacheAsJSON exports the entire cache as a single JSON object
func (c *Cache) ExportCacheAsJSON() (map[string]interface{}, error) {
	prompts, err := c.GetPrompts()
	if err != nil {
		return nil, err
	}

	categories, err := c.GetCategories()
	if err != nil {
		return nil, err
	}

	tags, err := c.GetTags()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"prompts":      prompts,
		"categories":   categories,
		"tags":         tags,
		"exported_at":  time.Now().Format(time.RFC3339),
		"cache_version": 1,
	}, nil
}

// PopulateFromJSON populates the cache from JSON data
func (c *Cache) PopulateFromJSON(data map[string]interface{}) error {
	// Clear existing data
	c.db.Exec("DELETE FROM prompts")
	c.db.Exec("DELETE FROM categories")
	c.db.Exec("DELETE FROM tags")

	// Import prompts
	if promptsData, ok := data["prompts"].([]interface{}); ok {
		for _, p := range promptsData {
			promptJSON, _ := json.Marshal(p)
			var prompt Prompt
			json.Unmarshal(promptJSON, &prompt)
			c.SavePrompt(prompt)
		}
	}

	// Import categories
	if categoriesData, ok := data["categories"].([]interface{}); ok {
		for _, cat := range categoriesData {
			catJSON, _ := json.Marshal(cat)
			var category Category
			json.Unmarshal(catJSON, &category)
			c.SaveCategory(category)
		}
	}

	// Import tags
	if tagsData, ok := data["tags"].([]interface{}); ok {
		for _, t := range tagsData {
			tagJSON, _ := json.Marshal(t)
			var tag Tag
			json.Unmarshal(tagJSON, &tag)
			c.SaveTag(tag)
		}
	}

	return nil
}
