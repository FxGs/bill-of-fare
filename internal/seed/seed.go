package seed

import (
	"database/sql"
	"fmt"

	"gopkg.in/yaml.v3"
)

type MenuFile struct {
	Categories []Category `yaml:"categories"`
}

type Category struct {
	Name  string `yaml:"name"`
	Items []Item `yaml:"items"`
}

type Item struct {
	Name     string    `yaml:"name"`
	Variants []Variant `yaml:"variants"`
}

type Variant struct {
	Name  string `yaml:"name"`
	Price int    `yaml:"price"`
}

func IsMenuEmpty(database *sql.DB) (bool, error) {
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM menu_items").Scan(&count); err != nil {
		return false, fmt.Errorf("count menu items: %w", err)
	}
	return count == 0, nil
}

func SeedFromYAML(database *sql.DB, content []byte) error {
	var doc MenuFile
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return fmt.Errorf("parse menu yaml: %w", err)
	}
	return Seed(database, doc)
}

func Seed(database *sql.DB, doc MenuFile) error {
	tx, err := database.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, c := range doc.Categories {
		if _, err := tx.Exec("INSERT OR IGNORE INTO categories(name) VALUES(?)", c.Name); err != nil {
			return err
		}
		var cid int
		if err := tx.QueryRow("SELECT id FROM categories WHERE name = ?", c.Name).Scan(&cid); err != nil {
			return err
		}
		for _, it := range c.Items {
			for _, v := range it.Variants {
				_, err := tx.Exec(`INSERT OR IGNORE INTO menu_items(category_id, name, variant, price) VALUES(?,?,?,?)`, cid, it.Name, v.Name, v.Price)
				if err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit()
}
