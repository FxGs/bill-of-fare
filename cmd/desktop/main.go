package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"bill-of-fare/internal/assets"
	"bill-of-fare/internal/db"
	"bill-of-fare/internal/handlers"
	"bill-of-fare/internal/seed"
	"bill-of-fare/internal/services"
)

func main() {
	dbPath, err := desktopDBPath()
	if err != nil {
		log.Fatal(err)
	}
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := seedIfEmpty(database); err != nil {
		log.Fatal(err)
	}

	tpl := template.Must(template.ParseFS(assets.FS, "web/templates/*.html"))
	staticFS, _ := fs.Sub(assets.FS, "web/static")
	h := handlers.Handler{
		Tpl:      tpl,
		Menu:     services.MenuService{DB: database},
		Cart:     services.NewCartService(),
		Invoices: services.InvoiceService{DB: database},
		Settings: services.SettingsService{DB: database},
		Static:   http.FileServer(http.FS(staticFS)),
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	url := fmt.Sprintf("http://%s", listener.Addr().String())
	go func() {
		time.Sleep(300 * time.Millisecond)
		_ = openBrowser(url)
	}()

	log.Printf("Bill of Fare running at %s", url)
	log.Fatal(http.Serve(listener, h.Routes()))
}

func desktopDBPath() (string, error) {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir, err = os.Executable()
		if err != nil {
			return "", err
		}
		dir = filepath.Dir(dir)
	}
	appDir := filepath.Join(dir, "Bill of Fare")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "bill_of_fare.db"), nil
}

func seedIfEmpty(database *sql.DB) error {
	empty, err := seed.IsMenuEmpty(database)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	content, err := assets.FS.ReadFile("seed/menu.yaml")
	if err != nil {
		return err
	}
	if err := seed.SeedFromYAML(database, content); err != nil {
		return err
	}
	return nil
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
