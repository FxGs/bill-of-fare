package services

import (
	"testing"

	"bill-of-fare/internal/models"
)

func TestInvoiceServiceCreateGetListExportAndSales(t *testing.T) {
	database := openTestDB(t)
	invoices := InvoiceService{DB: database}

	next, err := invoices.NextNumber()
	if err != nil {
		t.Fatalf("NextNumber before create: %v", err)
	}
	if next != 1 {
		t.Fatalf("NextNumber before create = %d, want 1", next)
	}

	items := []models.CartItem{
		{MenuItem: models.MenuItem{ID: 1, Name: "Mix Veg", Variant: "Half", Price: 130}, Quantity: 2, Subtotal: 260},
		{MenuItem: models.MenuItem{ID: 2, Name: "Naan", Price: 30}, Quantity: 1, Subtotal: 30},
	}
	id, err := invoices.Create(items, 290)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	inv, err := invoices.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if inv.ID != id || inv.Total != 290 || len(inv.Items) != 2 {
		t.Fatalf("invoice = %+v, want id %d total 290 two items", inv, id)
	}
	if inv.Items[0].ItemName != "Mix Veg (Half)" || inv.Items[0].Quantity != 2 || inv.Items[0].Subtotal != 260 {
		t.Fatalf("first invoice item = %+v, want variant suffix and subtotal", inv.Items[0])
	}

	next, err = invoices.NextNumber()
	if err != nil {
		t.Fatalf("NextNumber after create: %v", err)
	}
	if next != id+1 {
		t.Fatalf("NextNumber after create = %d, want %d", next, id+1)
	}

	summaries, err := invoices.List(10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(summaries) != 1 || summaries[0].ID != id || summaries[0].ItemCount != 2 || summaries[0].Total != 290 {
		t.Fatalf("summaries = %+v, want invoice summary", summaries)
	}

	rows, err := invoices.ExportRows()
	if err != nil {
		t.Fatalf("ExportRows: %v", err)
	}
	if len(rows) != 2 || rows[0].InvoiceID != id || rows[0].InvoiceTotal != 290 {
		t.Fatalf("export rows = %+v, want two rows for invoice %d", rows, id)
	}

	today, err := invoices.TodaySales()
	if err != nil {
		t.Fatalf("TodaySales: %v", err)
	}
	if today.Count != 1 || today.Total != 290 {
		t.Fatalf("TodaySales = %+v, want count 1 total 290", today)
	}
}

func TestInvoiceServiceListDefaultLimitAndVariantSuffix(t *testing.T) {
	database := openTestDB(t)
	invoices := InvoiceService{DB: database}

	first, err := invoices.Create([]models.CartItem{
		{MenuItem: models.MenuItem{ID: 1, Name: "Tea", Price: 20}, Quantity: 1, Subtotal: 20},
	}, 20)
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}
	second, err := invoices.Create([]models.CartItem{
		{MenuItem: models.MenuItem{ID: 2, Name: "Coffee", Variant: "Large", Price: 50}, Quantity: 1, Subtotal: 50},
	}, 50)
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}

	summaries, err := invoices.List(0)
	if err != nil {
		t.Fatalf("List default limit: %v", err)
	}
	if len(summaries) != 2 || summaries[0].ID != second || summaries[1].ID != first {
		t.Fatalf("summaries = %+v, want newest first", summaries)
	}
	if got := variantSuffix(""); got != "" {
		t.Fatalf("variantSuffix blank = %q, want blank", got)
	}
	if got := variantSuffix("Large"); got != " (Large)" {
		t.Fatalf("variantSuffix Large = %q, want suffix", got)
	}
}
