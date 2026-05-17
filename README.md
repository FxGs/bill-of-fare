# Bill of Fare

Phase 1 lightweight POS web app built with Go + HTMX + SQLite.

## Run

```bash
go mod tidy
go run ./cmd/seed -db bill_of_fare.db -menu seed/menu.yaml
go run ./cmd/server
```

Open http://localhost:8080

## Features
- Menu display by category
- Session cart (in-memory by session cookie)
- Quantity updates/removal
- Invoice generation + print page
- Single binary friendly (embedded templates/static files)
