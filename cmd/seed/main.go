package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"bill-of-fare/internal/db"
	"gopkg.in/yaml.v3"
)

type menuFile struct { Categories []category `yaml:"categories"` }
type category struct { Name string `yaml:"name"`; Items []item `yaml:"items"` }
type item struct { Name string `yaml:"name"`; Variants []variant `yaml:"variants"` }
type variant struct { Name string `yaml:"name"`; Price int `yaml:"price"` }

func main() {
	menuPath := flag.String("menu", "seed/menu.yaml", "path to menu yaml")
	dbPath := flag.String("db", "bill_of_fare.db", "path to sqlite db")
	flag.Parse()

	database, err := db.Open(*dbPath)
	must(err)
	defer database.Close()

	b, err := os.ReadFile(*menuPath)
	must(err)
	var doc menuFile
	must(yaml.Unmarshal(b, &doc))

	must(seed(database, doc))
	fmt.Println("seed completed")
}

func seed(database *sql.DB, doc menuFile) error {
	tx, err := database.Begin(); if err != nil { return err }
	defer tx.Rollback()
	for _, c := range doc.Categories {
		res, err := tx.Exec("INSERT OR IGNORE INTO categories(name) VALUES(?)", c.Name); if err != nil { return err }
		_ = res
		var cid int
		if err := tx.QueryRow("SELECT id FROM categories WHERE name = ?", c.Name).Scan(&cid); err != nil { return err }
		for _, it := range c.Items {
			for _, v := range it.Variants {
				_, err := tx.Exec(`INSERT OR IGNORE INTO menu_items(category_id, name, variant, price) VALUES(?,?,?,?)`, cid, it.Name, v.Name, v.Price)
				if err != nil { return err }
			}
		}
	}
	return tx.Commit()
}

func must(err error) { if err != nil { panic(err) } }
