package seed

import (
	"testing"

	"bill-of-fare/internal/db"
)

func TestSeedFromYAMLIsIdempotent(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})

	empty, err := IsMenuEmpty(database)
	if err != nil {
		t.Fatalf("IsMenuEmpty before seed: %v", err)
	}
	if !empty {
		t.Fatal("IsMenuEmpty before seed = false, want true")
	}

	content := []byte(`
categories:
  - name: Starters
    items:
      - name: Soup
        variants:
          - name: Cup
            price: 80
          - name: Bowl
            price: 120
`)
	if err := SeedFromYAML(database, content); err != nil {
		t.Fatalf("SeedFromYAML first: %v", err)
	}
	if err := SeedFromYAML(database, content); err != nil {
		t.Fatalf("SeedFromYAML second: %v", err)
	}

	empty, err = IsMenuEmpty(database)
	if err != nil {
		t.Fatalf("IsMenuEmpty after seed: %v", err)
	}
	if empty {
		t.Fatal("IsMenuEmpty after seed = true, want false")
	}

	var categoryCount, itemCount int
	if err := database.QueryRow("SELECT COUNT(*) FROM categories").Scan(&categoryCount); err != nil {
		t.Fatalf("count categories: %v", err)
	}
	if err := database.QueryRow("SELECT COUNT(*) FROM menu_items").Scan(&itemCount); err != nil {
		t.Fatalf("count menu items: %v", err)
	}
	if categoryCount != 1 || itemCount != 2 {
		t.Fatalf("counts = categories %d items %d, want 1 and 2", categoryCount, itemCount)
	}
}

func TestSeedFromYAMLRejectsInvalidYAML(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Fatalf("close db: %v", err)
		}
	})

	if err := SeedFromYAML(database, []byte("categories: [")); err == nil {
		t.Fatal("SeedFromYAML invalid YAML succeeded, want error")
	}
}

func TestSeedReportsDatabaseErrors(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	if _, err := IsMenuEmpty(database); err == nil {
		t.Fatal("IsMenuEmpty on closed db succeeded, want error")
	}
	if err := Seed(database, MenuFile{Categories: []Category{{Name: "Closed"}}}); err == nil {
		t.Fatal("Seed on closed db succeeded, want error")
	}
}
