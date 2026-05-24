package db

import (
	"path/filepath"
	"testing"
)

func TestOpenAppliesSchemaAndDefaultSettings(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})

	for _, table := range []string{"categories", "menu_items", "invoices", "invoice_items", "app_settings"} {
		t.Run(table, func(t *testing.T) {
			var name string
			if err := database.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&name); err != nil {
				t.Fatalf("table %s missing: %v", table, err)
			}
			if name != table {
				t.Fatalf("table name = %q, want %q", name, table)
			}
		})
	}

	var restaurantName string
	if err := database.QueryRow("SELECT value FROM app_settings WHERE key = 'restaurant_name'").Scan(&restaurantName); err != nil {
		t.Fatalf("default restaurant name missing: %v", err)
	}
	if restaurantName != "Bill of Fare" {
		t.Fatalf("restaurant name = %q, want Bill of Fare", restaurantName)
	}

	for _, column := range []string{"available", "best_seller"} {
		t.Run("menu_items_"+column, func(t *testing.T) {
			var count int
			if err := database.QueryRow("SELECT COUNT(*) FROM pragma_table_info('menu_items') WHERE name = ?", column).Scan(&count); err != nil {
				t.Fatalf("inspect menu_items column %s: %v", column, err)
			}
			if count != 1 {
				t.Fatalf("menu_items column %s count = %d, want 1", column, count)
			}
		})
	}
}

func TestOpenReturnsErrorForInvalidPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "bill_of_fare.db")
	if database, err := Open(path); err == nil {
		_ = database.Close()
		t.Fatal("Open invalid path succeeded, want error")
	}
}
