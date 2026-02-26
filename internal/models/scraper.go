package models

import (
	"database/sql"
)

// ScraperItem is a placeholder representation of what the scraper might gather.
type ScraperItem struct {
	ID    int
	Title string
	Value string
}

// ScraperModel wraps a database connection pool for the scraper specifically.
type ScraperModel struct {
	DB *sql.DB
}

// InitSchema creates the scraper tables if they don't exist.
func (m *ScraperModel) InitSchema() error {
	stmt := `
	CREATE TABLE IF NOT EXISTS scraped_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		value TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := m.DB.Exec(stmt)
	return err
}

// Insert adds a new item to the scraper DB.
func (m *ScraperModel) Insert(title, value string) (int, error) {
	stmt := `INSERT INTO scraped_items (title, value, created_at)
	VALUES(?, ?, CURRENT_TIMESTAMP) RETURNING id`

	var id int
	err := m.DB.QueryRow(stmt, title, value).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Latest returns the most recent scraped items.
func (m *ScraperModel) Latest(limit int) ([]*ScraperItem, error) {
	stmt := `SELECT id, title, value FROM scraped_items
	ORDER BY created_at DESC LIMIT ?`

	rows, err := m.DB.Query(stmt, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*ScraperItem

	for rows.Next() {
		e := &ScraperItem{}
		err = rows.Scan(&e.ID, &e.Title, &e.Value)
		if err != nil {
			return nil, err
		}
		items = append(items, e)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
