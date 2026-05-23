package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const schemaSQL = `
PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS menu_items (id INTEGER PRIMARY KEY, category_id INTEGER NOT NULL, name TEXT NOT NULL, variant TEXT, price INTEGER NOT NULL, available INTEGER NOT NULL DEFAULT 1, best_seller INTEGER NOT NULL DEFAULT 0, UNIQUE(category_id, name, variant), FOREIGN KEY(category_id) REFERENCES categories(id));
CREATE TABLE IF NOT EXISTS invoices (id INTEGER PRIMARY KEY, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, total INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS invoice_items (id INTEGER PRIMARY KEY, invoice_id INTEGER NOT NULL, item_name TEXT NOT NULL, quantity INTEGER NOT NULL, unit_price INTEGER NOT NULL, subtotal INTEGER NOT NULL, FOREIGN KEY(invoice_id) REFERENCES invoices(id));
CREATE TABLE IF NOT EXISTS app_settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);
INSERT OR IGNORE INTO app_settings(key, value) VALUES ('restaurant_name', 'Bill of Fare');
`

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := ensureMenuItemColumns(db); err != nil {
		return nil, err
	}
	return db, nil
}

func ensureMenuItemColumns(db *sql.DB) error {
	columns, err := tableColumns(db, "menu_items")
	if err != nil {
		return fmt.Errorf("inspect menu item columns: %w", err)
	}
	if !columns["available"] {
		if _, err := db.Exec("ALTER TABLE menu_items ADD COLUMN available INTEGER NOT NULL DEFAULT 1"); err != nil {
			return fmt.Errorf("add available column: %w", err)
		}
	}
	if !columns["best_seller"] {
		if _, err := db.Exec("ALTER TABLE menu_items ADD COLUMN best_seller INTEGER NOT NULL DEFAULT 0"); err != nil {
			return fmt.Errorf("add best seller column: %w", err)
		}
	}
	return nil
}

func tableColumns(db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, pk int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		columns[name] = true
	}
	return columns, rows.Err()
}
