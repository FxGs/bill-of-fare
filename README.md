# Bill of Fare

Phase 1 lightweight POS web app built with Go + HTMX + SQLite.

## Run

```bash
go mod tidy
go run ./cmd/seed -db bill_of_fare.db -menu seed/menu.yaml
go run ./cmd/server
```

Open http://localhost:8080

## Dev

Run with restart-on-change while editing Go, templates, CSS, schema, or seed files:

```bash
./scripts/dev.sh
```

Optional environment variables:

```bash
DB_PATH=/tmp/bill_of_fare.db MENU_PATH=seed/menu.yaml PORT=8080 ./scripts/dev.sh
```

## Windows Desktop Build

Build a one-click Windows executable:

```bash
./scripts/package-windows.sh
```

The output is `dist/windows/BillOfFare.exe`. On Windows, double-clicking it starts the local POS and opens it in the default browser. The app stores its database in `%AppData%\Bill of Fare\bill_of_fare.db` and seeds the menu automatically on first run.

## Features
- Menu display by category
- Session cart (in-memory by session cookie)
- Quantity updates/removal
- Invoice generation + print page
- Single binary friendly (embedded templates/static files)
