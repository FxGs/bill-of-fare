# Bill of Fare — Phase 1 Codex Plan

## Goal

Build a lightweight restaurant POS web app that:

- loads menu data from SQLite
- displays categorized dishes
- allows adding dishes to a cart
- generates a printable invoice
- runs locally as a single Go binary

No authentication.  
No payment integration.  
No admin panel.  
No cloud deployment.

---

# Tech Stack

## Backend

- Go 1.24+
- Gin (or standard `net/http`)

## Frontend

- HTMX
- Go HTML templates
- Bootstrap 5

## Database

- SQLite

## Packaging

- Single executable binary
- Embedded templates/static assets using Go `embed`

---

# Architecture

```text
Browser
  ↓
HTMX Requests
  ↓
Go Server
  ↓
SQLite
```

Server returns HTML fragments instead of JSON APIs.

---

# Project Structure

```text
bill-of-fare/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   ├── db/
│   ├── models/
│   ├── handlers/
│   ├── services/
│   └── templates/
│
├── web/
│   ├── templates/
│   └── static/
│
├── seed/
│   └── menu.yaml
│
├── database/
│   └── schema.sql
│
├── go.mod
└── README.md
```

---

# Phase 1 Features

## 1. Menu Display

Display categories:
- Roll
- Pizza
- Momos
- Mocktail
- etc.

Each item shows:
- name
- variant (optional)
- price
- add button

---

## 2. Cart System

User can:
- add items
- increase quantity
- remove items
- view running total

Cart stored:
- in-memory session
OR
- lightweight cookie/session store

No user accounts required.

---

## 3. Invoice Generation

Generate invoice page:
- item list
- quantities
- subtotals
- grand total
- timestamp

Must support browser printing.

---

# Menu Data Format

Use YAML as the source of truth.

Example:

```yaml
categories:
  - name: Roll
    items:
      - name: Chicken Roll
        variants:
          - name: Regular
            price: 80
          - name: Cheese
            price: 100
```

---

# Database Schema

## categories

```sql
CREATE TABLE categories (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL
);
```

## menu_items

```sql
CREATE TABLE menu_items (
  id INTEGER PRIMARY KEY,
  category_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  variant TEXT,
  price INTEGER NOT NULL,
  FOREIGN KEY(category_id) REFERENCES categories(id)
);
```

## invoices

```sql
CREATE TABLE invoices (
  id INTEGER PRIMARY KEY,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  total INTEGER NOT NULL
);
```

## invoice_items

```sql
CREATE TABLE invoice_items (
  id INTEGER PRIMARY KEY,
  invoice_id INTEGER NOT NULL,
  item_name TEXT NOT NULL,
  quantity INTEGER NOT NULL,
  unit_price INTEGER NOT NULL,
  subtotal INTEGER NOT NULL,
  FOREIGN KEY(invoice_id) REFERENCES invoices(id)
);
```

---

# Seeder Requirements

Create a YAML-to-SQLite seeder.

Responsibilities:
- parse `menu.yaml`
- populate categories
- populate menu_items
- avoid duplicate inserts

Seeder runs automatically if DB is empty.

---

# Routes

## Page Routes

```text
GET  /
GET  /invoice/{id}
```

## HTMX Routes

```text
POST /cart/add
POST /cart/remove
POST /cart/update
GET  /cart
POST /checkout
```

HTMX endpoints return HTML partials only.

---

# HTMX Requirements

Use:
- `hx-post`
- `hx-get`
- `hx-target`
- `hx-swap`

Avoid:
- React
- JSON APIs
- frontend state libraries

Example pattern:

```html
<button
  hx-post="/cart/add"
  hx-target="#cart"
  hx-swap="outerHTML">
  Add
</button>
```

---

# Templates

## Required Templates

```text
layout.html
index.html
menu.html
cart.html
invoice preview modal partial
```

Use:
- template inheritance/layouts
- reusable partials

---

# UI Requirements

## Layout

Left:
- categories/menu

Right:
- cart

Top:
- app title ("Bill of Fare")

---

# Styling

Use Bootstrap only.

Avoid:
- Tailwind
- frontend build pipelines
- npm

Keep:
- zero JS build tooling

---

# Printing Requirements

Invoice page must:
- support browser print dialog
- use print CSS
- hide navigation/buttons when printing

---

# Asset Embedding

Embed:
- HTML templates
- CSS
- HTMX script

using Go `embed.FS`.

Final deliverable should run as:

```bash
./bill-of-fare
```

without external assets.

---

# Non-Goals (Phase 1)

Do NOT implement:
- authentication
- payments
- GST
- kitchen display
- inventory
- table management
- admin dashboard
- multi-user support
- websocket updates

---

# Phase 2 Ideas

Future enhancements:
- search/filter
- modifiers/add-ons
- half/full pricing
- sales history
- dashboard
- PDF receipts
- QR menu
- inventory tracking

---

# Coding Standards

- keep handlers thin
- separate DB logic into services/repositories
- use prepared statements
- avoid global mutable state
- prefer stdlib where possible
- keep dependencies minimal

---

# Deliverables

Codex should produce:

1. Complete Go project
2. SQLite schema
3. YAML parser/seeder
4. Embedded templates/static assets
5. HTMX cart flow
6. Printable invoice
7. README with build/run instructions

---

# Build Command

```bash
go build -o bill-of-fare ./cmd/server
```

---

# Expected Final UX

1. Launch executable
2. Open localhost page
3. Browse menu
4. Add dishes
5. Generate invoice
6. Print receipt
