package main

import (
	"flag"
	"fmt"
	"os"

	"bill-of-fare/internal/db"
	"bill-of-fare/internal/seed"
)

func main() {
	menuPath := flag.String("menu", "seed/menu.yaml", "path to menu yaml")
	dbPath := flag.String("db", "bill_of_fare.db", "path to sqlite db")
	flag.Parse()

	database, err := db.Open(*dbPath)
	must(err)
	defer database.Close()

	content, err := os.ReadFile(*menuPath)
	must(err)
	must(seed.SeedFromYAML(database, content))
	fmt.Println("seed completed")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
