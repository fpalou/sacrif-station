package models

import (
	"database/sql"
	"time"
)

// Entry defines the core flexible content unit of Sacrif Station.
type Entry struct {
	ID        int
	Title     string
	Type      string // e.g., "thought", "book", "game", "link", "log", "anime"
	Content   string
	URL       string // Optional
	CreatedAt time.Time
}

// EntryModel wraps a database connection pool.
type EntryModel struct {
	DB *sql.DB
}

// InitSchema creates the entries table if it doesn't exist.
func (m *EntryModel) InitSchema() error {
	stmt := `
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		type TEXT NOT NULL,
		content TEXT,
		url TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := m.DB.Exec(stmt)
	return err
}

// Insert adds a new entry to the database.
func (m *EntryModel) Insert(title, entryType, content, url string) (int, error) {
	stmt := `INSERT INTO entries (title, type, content, url, created_at)
	VALUES(?, ?, ?, ?, CURRENT_TIMESTAMP) RETURNING id`

	var id int
	err := m.DB.QueryRow(stmt, title, entryType, content, url).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Latest returns the most recent entries of ALL types.
func (m *EntryModel) Latest(limit int) ([]*Entry, error) {
	stmt := `SELECT id, title, type, content, url, created_at FROM entries
	ORDER BY created_at DESC LIMIT ?`
	return m.queryEntries(stmt, limit)
}

// LatestByType returns the most recent entries matching a specific type.
func (m *EntryModel) LatestByType(entryType string, limit int) ([]*Entry, error) {
	stmt := `SELECT id, title, type, content, url, created_at FROM entries
	WHERE type = ? ORDER BY created_at DESC LIMIT ?`
	return m.queryEntries(stmt, entryType, limit)
}

// LatestExcluded returns the most recent entries EXCLUDING a specific type.
func (m *EntryModel) LatestExcluded(excludedType string, limit int) ([]*Entry, error) {
	stmt := `SELECT id, title, type, content, url, created_at FROM entries
	WHERE type != ? ORDER BY created_at DESC LIMIT ?`
	return m.queryEntries(stmt, excludedType, limit)
}

// Helper method to execute a query returning multiple entries
func (m *EntryModel) queryEntries(stmt string, args ...any) ([]*Entry, error) {
	rows, err := m.DB.Query(stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*Entry

	for rows.Next() {
		e := &Entry{}
		err = rows.Scan(&e.ID, &e.Title, &e.Type, &e.Content, &e.URL, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// Count returns the total number of entries
func (m *EntryModel) Count() (int, error) {
	var count int
	err := m.DB.QueryRow(`SELECT COUNT(*) FROM entries`).Scan(&count)
	return count, err
}
