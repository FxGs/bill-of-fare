package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const schemaSQL = `
PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS menu_items (id INTEGER PRIMARY KEY, category_id INTEGER NOT NULL, name TEXT NOT NULL, variant TEXT, price INTEGER NOT NULL, UNIQUE(category_id, name, variant), FOREIGN KEY(category_id) REFERENCES categories(id));
CREATE TABLE IF NOT EXISTS invoices (id INTEGER PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, total INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS invoice_items (id INTEGER PRIMARY KEY, invoice_id INTEGER NOT NULL, item_name TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price INTEGER NOT NULL, subtotal INTEGER NOT NULL, FOREIGN KEY(invoice_id) REFERENCES invoices(id));
`

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil { return nil, fmt.Errorf("open sqlite: %w", err) }
	if err := db.Ping(); err != nil { return nil, fmt.Errorf("ping sqlite: %w", err) }
	if _, err := db.Exec(schemaSQL); err != nil { return nil, fmt.Errorf("apply schema: %w", err) }
	return db, nil
}
