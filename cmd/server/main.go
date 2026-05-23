package main

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"bill-of-fare/internal/assets"
	"bill-of-fare/internal/build"
	"bill-of-fare/internal/db"
	"bill-of-fare/internal/handlers"
	"bill-of-fare/internal/services"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "bill_of_fare.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	tpl := template.Must(template.ParseFS(assets.FS, "web/templates/*.html"))
	staticFS, _ := fs.Sub(assets.FS, "web/static")
	h := handlers.Handler{Tpl: tpl, Menu: services.MenuService{DB: database}, Cart: services.NewCartService(), Invoices: services.InvoiceService{DB: database}, Settings: services.SettingsService{DB: database}, Version: build.Version, Static: http.FileServer(http.FS(staticFS))}
	addr := host + ":" + port
	log.Println("server listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, h.Routes()))
}
