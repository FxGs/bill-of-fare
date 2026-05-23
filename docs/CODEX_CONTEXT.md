# Bill of Fare Codex Context

This note is for future Codex sessions working in this repo. Keep it short, factual, and updated when behavior changes.

## Project Shape

- Lightweight local restaurant POS built in Go with standard `net/http`, Go templates, HTMX, CSS, and SQLite.
- Server entrypoint: `cmd/server/main.go`.
- Seeder entrypoint: `cmd/seed/main.go`.
- Database schema: `database/schema.sql`.
- Menu seed source: `seed/menu.yaml`.
- The root `menu.yaml` was the source menu provided by the user; `seed/menu.yaml` is the app seed format.
- Runtime DB defaults to `bill_of_fare.db`, which is ignored by git via `*.db`.

## Asset Layout

- Editable template/static source lives in `web/`.
- Embedded app assets live in `internal/assets/web/` and are served via Go `embed` from `internal/assets/assets.go`.
- Keep `web/` and `internal/assets/web/` in sync after template or CSS changes.
- Quick sync check:

```bash
diff -qr web internal/assets/web
```

## Local Development

- Run the normal server:

```bash
go run ./cmd/server
```

- Run with hot reload:

```bash
./scripts/dev.sh
```

- Useful environment overrides:

```bash
DB_PATH=/tmp/bill_of_fare.db MENU_PATH=seed/menu.yaml HOST=127.0.0.1 PORT=8080 ./scripts/dev.sh
```

- `scripts/dev.sh` seeds the DB if missing, builds a temporary binary, starts it, and restarts when files in `cmd`, `internal`, `web`, `database`, or `seed` change.
- `cmd/server` defaults to `HOST=127.0.0.1` and `PORT=8080`.
- Avoid killing the dev server unless the user asks or the port is actually blocked.

## Windows Packaging

- `cmd/desktop` is the Windows-friendly one-click entrypoint.
- It stores the DB in the current Windows user's config directory under `Bill of Fare/bill_of_fare.db`.
- It seeds the embedded `internal/assets/seed/menu.yaml` on first run.
- It binds to a free `127.0.0.1` port and opens the default browser automatically.
- Build-time version lives in `internal/build.Version` and is stamped with `-ldflags "-X bill-of-fare/internal/build.Version=<version>"`.
- POS/admin headers render that build version.
- The executable is self-contained for the app code, embedded web assets, schema, starter menu, and SQLite driver. Live data is intentionally stored outside the binary in the user's app-data directory.
- Build it with:

```bash
./scripts/package-windows.sh
```

- Output goes to `dist/windows/BillOfFare.exe`; `dist/` is intentionally gitignored.
- GitHub Actions releases are tag-driven. Push a tag like `1.2.0` or `v1.2.0` to build a Windows `x64` ZIP and attach it to a GitHub Release. The build strips a leading `v` from the version shown in the app header.

## Current Routes

- `GET /` renders the POS.
- `GET /menu?category_id=N` renders the menu fragment. `category_id=0` or empty shows all categories.
- `POST /cart/add`, `/cart/qty`, `/cart/remove`, and `GET /cart` update/render the cart fragment.
- `GET /sales` renders the sales summary modal.
- `GET /admin/invoices/export` exports invoice line items as CSV.
- `POST /admin/settings/restaurant-name` updates the restaurant name printed on invoices.
- `POST /invoice/create` creates an invoice, records session sales, clears the cart, and redirects to `/invoice?id=N`.
- `GET /invoice?id=N` renders a printable invoice.
- `GET /admin` renders menu admin.
- `POST /admin/categories/create`, `/admin/categories/delete`, `/admin/items/create`, `/admin/items/update`, `/admin/items/delete` mutate menu data.

## POS Behavior

- Categories on the POS filter the visible menu instead of scrolling.
- POS menu display groups variants under one dish card; multi-variant dishes open a variant chooser modal before adding to cart.
- Cart state is in memory per `session_id` cookie.
- Invoice numbers come from `MAX(invoices.id) + 1`; the visible order number should not be hardcoded.
- Sales modal shows:
  - today’s persisted invoice sales from SQLite
  - this session’s completed sales from in-memory cart service
  - current open-cart total

## Admin Behavior

- Admin is reachable from the top nav at `/admin`.
- Admin uses the same dark visual language as the POS.
- Add item, category management, category deletion, and item deletion use modal flows.
- Category deletion is blocked in service code while any menu items still use that category.
- Category management opens in a `Manage Categories` modal from the admin action strip.
- New category creation is at the top of the `Manage Categories` modal.
- Menu rows support category, item name, variant, price, save, and delete.
- Save uses a floppy disk icon; delete uses a trash icon.
- Price inputs hide browser number spinners.
- The menu item table has client-side search and category filtering, plus a live visible-row count and empty state.
- Past invoices are available from the admin action strip, with printable invoice links and CSV export.
- Receipt settings are available from the admin action strip; `restaurant_name` is stored in `app_settings`.
- Mobile admin is acceptable but not the primary polish target right now.

## Data Model Notes

- `categories.name` is unique.
- `menu_items` has a uniqueness constraint on `(category_id, name, variant)`.
- `invoices.created_at` defaults to SQLite `CURRENT_TIMESTAMP`.
- `app_settings` stores simple app configuration, including `restaurant_name`.
- `TodaySales()` uses SQLite local-date comparison:

```sql
DATE(created_at, 'localtime') = DATE('now', 'localtime')
```

## Git And GitHub Setup

- This repo has `.envrc` setting:

```bash
export GH_CONFIG_DIR="$HOME/.config/gh-fxgs"
```

- The user authenticated GitHub CLI for this repo-specific config.
- Use `env GH_CONFIG_DIR=/Users/abhishek/.config/gh-fxgs gh ...` in automation if direnv is not loaded in the shell.
- Git commit identity is configured per repo for the user’s preferred account.
- Branch names should use the `codex/` prefix unless the user asks otherwise.

## Verification Checklist

- Run:

```bash
go test ./...
npm run test:ui
diff -qr web internal/assets/web
```

- For frontend/admin changes, verify in the in-app browser at `http://localhost:8080` or `http://localhost:8080/admin`.
- Browser UI tests live in `tests/ui` and use Playwright. The config starts `cmd/server` against a seeded throwaway SQLite DB on port `8090` by default.
- When testing admin filters, useful checks:
  - search narrows rows
  - impossible search shows `No matching menu items`
  - category dropdown narrows to only that category
  - table header and row columns stay aligned

## Open Context

- Current work is admin/menu-editing improvements.
- The user has been actively reviewing visual details in the in-app browser and prefers small, practical UI refinements over big redesigns.
- Do not revert user or generated changes you did not make. Check `git status --short --branch` before committing.
