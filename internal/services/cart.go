package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"bill-of-fare/internal/models"
)

type cart struct {
	items map[string]*models.CartItem
}

type CartService struct {
	mu    sync.RWMutex
	carts map[string]*cart
}

func NewCartService() *CartService { return &CartService{carts: map[string]*cart{}} }

func (s *CartService) SessionID(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("session_id"); err == nil && c.Value != "" { return c.Value }
	b := make([]byte, 16); _, _ = rand.Read(b); id := hex.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{Name: "session_id", Value: id, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Expires: time.Now().Add(24 * time.Hour)})
	return id
}

func (s *CartService) ensure(session string) *cart {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.carts[session]; !ok { s.carts[session] = &cart{items: map[string]*models.CartItem{}} }
	return s.carts[session]
}

func (s *CartService) Add(session string, item models.MenuItem) {
	c := s.ensure(session)
	key := fmt.Sprintf("%d", item.ID)
	s.mu.Lock(); defer s.mu.Unlock()
	if existing, ok := c.items[key]; ok { existing.Quantity++ ; existing.Subtotal = existing.Quantity * existing.MenuItem.Price; return }
	c.items[key] = &models.CartItem{Key:key, MenuItem:item, Quantity:1, Subtotal:item.Price}
}

func (s *CartService) ChangeQty(session, key string, delta int) {
	c := s.ensure(session)
	s.mu.Lock(); defer s.mu.Unlock()
	if it, ok := c.items[key]; ok { it.Quantity += delta; if it.Quantity <=0 { delete(c.items,key); return }; it.Subtotal = it.Quantity * it.MenuItem.Price }
}

func (s *CartService) Remove(session, key string) { c:=s.ensure(session); s.mu.Lock(); defer s.mu.Unlock(); delete(c.items,key) }

func (s *CartService) Snapshot(session string) ([]models.CartItem, int) {
	c := s.ensure(session)
	s.mu.RLock(); defer s.mu.RUnlock()
	items := make([]models.CartItem,0,len(c.items)); total:=0
	for _, it := range c.items { v := *it; items=append(items,v); total+=v.Subtotal }
	return items,total
}

func (s *CartService) Clear(session string) { c:=s.ensure(session); s.mu.Lock(); defer s.mu.Unlock(); c.items = map[string]*models.CartItem{} }

type InvoiceService struct { DB *sql.DB }

func (s InvoiceService) Create(items []models.CartItem, total int) (int, error) {
	tx, err := s.DB.Begin()
	if err != nil { return 0, err }
	defer tx.Rollback()
	res, err := tx.Exec("INSERT INTO invoices(total) VALUES(?)", total)
	if err != nil { return 0, err }
	invID, _ := res.LastInsertId()
	for _, it := range items {
		_, err := tx.Exec(`INSERT INTO invoice_items(invoice_id, item_name, quantity, unit_price, subtotal) VALUES(?,?,?,?,?)`, invID, it.MenuItem.Name+variantSuffix(it.MenuItem.Variant), it.Quantity, it.MenuItem.Price, it.Subtotal)
		if err != nil { return 0, err }
	}
	if err := tx.Commit(); err != nil { return 0, err }
	return int(invID), nil
}

func (s InvoiceService) Get(id int) (models.Invoice, error) {
	var inv models.Invoice
	if err := s.DB.QueryRow("SELECT id, created_at, total FROM invoices WHERE id = ?", id).Scan(&inv.ID, &inv.CreatedAt, &inv.Total); err != nil { return inv, err }
	rows, err := s.DB.Query("SELECT item_name, quantity, unit_price, subtotal FROM invoice_items WHERE invoice_id = ?", id)
	if err != nil { return inv, err }
	defer rows.Close()
	for rows.Next() { var it models.InvoiceItem; if err := rows.Scan(&it.ItemName,&it.Quantity,&it.UnitPrice,&it.Subtotal); err != nil { return inv, err }; inv.Items=append(inv.Items,it) }
	return inv, rows.Err()
}

func variantSuffix(v string) string { if v=="" { return ""}; return " ("+v+")" }
