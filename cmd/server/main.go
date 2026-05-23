package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"bill-of-fare/internal/assets"
	"bill-of-fare/internal/db"
	"bill-of-fare/internal/handlers"
	"bill-of-fare/internal/services"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "bill_of_fare.db"
	}
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	tpl := template.Must(template.ParseFS(assets.FS, "web/templates/*.html"))
	staticFS, _ := fs.Sub(assets.FS, "web/static")
	h := handlers.Handler{Tpl: tpl, Menu: services.MenuService{DB: database}, Cart: services.NewCartService(), Invoices: services.InvoiceService{DB: database}, Static: http.FileServer(http.FS(staticFS))}
	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", h.Routes()))
}
