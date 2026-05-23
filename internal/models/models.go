package models

import "time"

type Category struct {
	ID    int
	Name  string
	Items []MenuItem
}

type MenuItem struct {
	ID           int
	CategoryID   int
	CategoryName string
	Name         string
	Variant      string
	Price        int
}

type CartItem struct {
	Key      string
	MenuItem MenuItem
	Quantity int
	Subtotal int
}

type Invoice struct {
	ID        int
	CreatedAt time.Time
	Items     []InvoiceItem
	Total     int
}

type InvoiceItem struct {
	ItemName  string
	Quantity  int
	UnitPrice int
	Subtotal  int
}

type SalesSummary struct {
	Count int
	Total int
}
