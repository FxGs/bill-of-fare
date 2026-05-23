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

The executable is self-contained for the application code: it includes the server, web UI, database schema, starter menu, and SQLite driver. It does not require Go, a separate web server, or seed files next to the executable. Live restaurant data remains outside the binary in the user's app-data folder by design.

To stamp a local build with a version:

```bash
VERSION=1.2.0 ./scripts/package-windows.sh
```

## Releases

Create a GitHub Release by pushing a tag in `x.y.z` or `vx.y.z` format:

```bash
git tag v1.2.0
git push origin v1.2.0
```

GitHub Actions builds a Windows `x64` ZIP file, stamps the app header with the tag version, and attaches the package to the release.

## Features
- Menu display by category
- Session cart (in-memory by session cookie)
- Quantity updates/removal
- Invoice generation + print page
- Single binary friendly (embedded templates/static files)
