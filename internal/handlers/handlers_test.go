package handlers

import (
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"bill-of-fare/internal/assets"
	"bill-of-fare/internal/db"
	"bill-of-fare/internal/models"
	"bill-of-fare/internal/services"
)

func newTestHandler(t *testing.T) Handler {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	database.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Fatalf("close test db: %v", err)
		}
	})
	tpl := template.Must(template.ParseFS(assets.FS, "web/templates/*.html"))
	staticFS, err := fs.Sub(assets.FS, "web/static")
	if err != nil {
		t.Fatalf("static fs: %v", err)
	}
	return Handler{
		Tpl:      tpl,
		Menu:     services.MenuService{DB: database},
		Cart:     services.NewCartService(),
		Invoices: services.InvoiceService{DB: database},
		Settings: services.SettingsService{DB: database},
		Version:  "test",
		Static:   http.FileServer(http.FS(staticFS)),
	}
}

func formRequest(method, target string, values url.Values) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d; body %s", rec.Code, want, rec.Body.String())
	}
}

func assertContains(t *testing.T, body string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q:\n%s", want, body)
		}
	}
}

func TestIndexAndMenuPaneRenderGroupedVariants(t *testing.T) {
	h := newTestHandler(t)
	if err := h.Menu.CreateCategory("Biryani"); err != nil {
		t.Fatalf("CreateCategory Biryani: %v", err)
	}
	if err := h.Menu.CreateCategory("Chinese"); err != nil {
		t.Fatalf("CreateCategory Chinese: %v", err)
	}
	cats, _ := h.Menu.ListCategories()
	biryaniID := cats[0].ID
	chineseID := cats[1].ID
	if cats[0].Name != "Biryani" {
		biryaniID, chineseID = cats[1].ID, cats[0].ID
	}
	if err := h.Menu.CreateMenuItem(biryaniID, "", "Chicken Biryani", "Half", 160); err != nil {
		t.Fatalf("CreateMenuItem biryani half: %v", err)
	}
	if err := h.Menu.CreateMenuItem(biryaniID, "", "Chicken Biryani", "Full", 260); err != nil {
		t.Fatalf("CreateMenuItem biryani full: %v", err)
	}
	if err := h.Menu.CreateMenuItem(chineseID, "", "Chicken Chilli", "", 180); err != nil {
		t.Fatalf("CreateMenuItem chinese: %v", err)
	}

	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	assertStatus(t, rec, http.StatusOK)
	body := rec.Body.String()
	assertContains(t, body, "Bill of Fare", "POS vtest", "Chicken Biryani", "2 variants", "₹160 - ₹260", "Chicken Chilli", "Create Order")

	rec = httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/menu?category_id="+strconv.Itoa(biryaniID), nil))
	assertStatus(t, rec, http.StatusOK)
	body = rec.Body.String()
	assertContains(t, body, "Chicken Biryani", "active")
	if strings.Contains(body, "Chicken Chilli") {
		t.Fatalf("filtered menu should not render Chinese item:\n%s", body)
	}
}

func TestCartRoutesMutateSessionCart(t *testing.T) {
	h := newTestHandler(t)
	if err := h.Menu.CreateCategory("Barbeque"); err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	cats, _ := h.Menu.ListCategories()
	if err := h.Menu.CreateMenuItem(cats[0].ID, "", "Paneer Tikka", "5pc", 180); err != nil {
		t.Fatalf("CreateMenuItem: %v", err)
	}
	items, _ := h.Menu.ListMenuItems()
	sessionCookie := &http.Cookie{Name: "session_id", Value: "cart-session"}

	rec := httptest.NewRecorder()
	req := formRequest(http.MethodPost, "/cart/add", url.Values{"item_id": {strconv.Itoa(items[0].ID)}})
	req.AddCookie(sessionCookie)
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusOK)
	assertContains(t, rec.Body.String(), "Paneer Tikka", "5pc", "₹180", "Create Order")

	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/cart/qty", url.Values{"key": {"1"}, "delta": {"1"}})
	req.AddCookie(sessionCookie)
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusOK)
	assertContains(t, rec.Body.String(), "₹360", ">2<")

	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/cart/remove", url.Values{"key": {"1"}})
	req.AddCookie(sessionCookie)
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusOK)
	assertContains(t, rec.Body.String(), "No items added yet", "disabled")
}

func TestSalesSummaryRendersCurrentSessionAndToday(t *testing.T) {
	h := newTestHandler(t)
	session := "sales-session"
	h.Cart.Add(session, models.MenuItem{ID: 3, Name: "Soup", Price: 80})
	h.Cart.RecordSale(session, 120)
	if _, err := h.Invoices.Create([]models.CartItem{
		{MenuItem: models.MenuItem{ID: 4, Name: "Noodles", Price: 150}, Quantity: 1, Subtotal: 150},
	}, 150); err != nil {
		t.Fatalf("Create invoice: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sales", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: session})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusOK)
	assertContains(t, rec.Body.String(), "Sales", "Today", "₹150", "This Session", "₹120", "Current Open Order", "₹80")
}

func TestAdminPageAndAdminMutationRoutes(t *testing.T) {
	h := newTestHandler(t)
	if err := h.Menu.CreateCategory("Starter"); err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	cats, _ := h.Menu.ListCategories()
	starterID := cats[0].ID
	if err := h.Menu.CreateMenuItem(starterID, "", "Soup", "", 80); err != nil {
		t.Fatalf("CreateMenuItem: %v", err)
	}

	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin", nil))
	assertStatus(t, rec, http.StatusOK)
	assertContains(t, rec.Body.String(), "Menu Admin", "Manage Categories", "Past Invoices", "Receipt Settings", "Soup")

	rec = httptest.NewRecorder()
	req := formRequest(http.MethodPost, "/admin/categories/create", url.Values{"name": {"Dessert"}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusSeeOther)
	if rec.Header().Get("Location") != "/admin" {
		t.Fatalf("category create redirect = %q, want /admin", rec.Header().Get("Location"))
	}

	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/admin/categories/create", url.Values{"name": {""}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusBadRequest)
	assertContains(t, rec.Body.String(), "category name is required")

	cats, _ = h.Menu.ListCategories()
	var dessertID int
	for _, c := range cats {
		if c.Name == "Dessert" {
			dessertID = c.ID
		}
	}
	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/admin/items/create", url.Values{"category_id": {strconv.Itoa(dessertID)}, "name": {"Panna Cotta"}, "variant": {""}, "price": {"120"}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusSeeOther)

	items, _ := h.Menu.ListMenuItems()
	var panna models.MenuItem
	for _, item := range items {
		if item.Name == "Panna Cotta" {
			panna = item
		}
	}
	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/admin/items/update", url.Values{"id": {strconv.Itoa(panna.ID)}, "category_id": {strconv.Itoa(dessertID)}, "name": {"Panna Cotta"}, "variant": {"Mini"}, "price": {"90"}, "available": {"on"}, "best_seller": {"on"}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusSeeOther)

	updated, err := h.Menu.GetMenuItem(panna.ID)
	if err != nil {
		t.Fatalf("GetMenuItem updated: %v", err)
	}
	if updated.Variant != "Mini" || updated.Price != 90 || !updated.BestSeller || !updated.Available {
		t.Fatalf("updated item = %+v, want Mini 90 available best seller", updated)
	}

	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/admin/items/delete", url.Values{"id": {strconv.Itoa(panna.ID)}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusSeeOther)

	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/admin/categories/delete", url.Values{"id": {strconv.Itoa(dessertID)}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusSeeOther)

	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/admin/categories/delete", url.Values{"id": {strconv.Itoa(starterID)}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusBadRequest)
	assertContains(t, rec.Body.String(), "delete or move 1 menu items")
}

func TestBestSellerPresetAndUnavailableItems(t *testing.T) {
	h := newTestHandler(t)
	if err := h.Menu.CreateCategory("Mains"); err != nil {
		t.Fatalf("CreateCategory Mains: %v", err)
	}
	if err := h.Menu.CreateCategory("Drinks"); err != nil {
		t.Fatalf("CreateCategory Drinks: %v", err)
	}
	cats, _ := h.Menu.ListCategories()
	var mainsID, drinksID int
	for _, c := range cats {
		if c.Name == "Mains" {
			mainsID = c.ID
		}
		if c.Name == "Drinks" {
			drinksID = c.ID
		}
	}
	if err := h.Menu.CreateMenuItem(mainsID, "", "Paneer Tikka", "", 180); err != nil {
		t.Fatalf("CreateMenuItem Paneer: %v", err)
	}
	if err := h.Menu.CreateMenuItem(drinksID, "", "Lassi", "", 70); err != nil {
		t.Fatalf("CreateMenuItem Lassi: %v", err)
	}
	items, _ := h.Menu.ListMenuItems()
	for _, item := range items {
		switch item.Name {
		case "Paneer Tikka":
			if err := h.Menu.UpdateMenuItem(item.ID, item.CategoryID, item.Name, item.Variant, item.Price, true, true); err != nil {
				t.Fatalf("mark best seller: %v", err)
			}
		case "Lassi":
			if err := h.Menu.UpdateMenuItem(item.ID, item.CategoryID, item.Name, item.Variant, item.Price, false, true); err != nil {
				t.Fatalf("mark unavailable: %v", err)
			}
		}
	}

	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	assertStatus(t, rec, http.StatusOK)
	body := rec.Body.String()
	assertContains(t, body, "Best Sellers", "Paneer Tikka", "Lassi", "Unavailable")

	rec = httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/menu?category_id="+strconv.Itoa(services.BestSellersCategoryID), nil))
	assertStatus(t, rec, http.StatusOK)
	body = rec.Body.String()
	assertContains(t, body, "Best Sellers", "Paneer Tikka", "Lassi", "Mains", "Unavailable")
}

func TestAdminSettingsAndInvoiceExportRoutes(t *testing.T) {
	h := newTestHandler(t)
	if _, err := h.Invoices.Create([]models.CartItem{
		{MenuItem: models.MenuItem{ID: 1, Name: "Coffee", Price: 50}, Quantity: 2, Subtotal: 100},
	}, 100); err != nil {
		t.Fatalf("Create invoice: %v", err)
	}

	rec := httptest.NewRecorder()
	req := formRequest(http.MethodPost, "/admin/settings/restaurant-name", url.Values{"restaurant_name": {"  Cafe Bills  "}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusSeeOther)
	if got := h.Settings.RestaurantName(); got != "Cafe Bills" {
		t.Fatalf("restaurant name = %q, want Cafe Bills", got)
	}

	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/admin/settings/restaurant-name", url.Values{"restaurant_name": {" "}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusBadRequest)
	assertContains(t, rec.Body.String(), "restaurant name is required")

	rec = httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/invoices/export", nil))
	assertStatus(t, rec, http.StatusOK)
	if got := rec.Header().Get("Content-Type"); got != "text/csv" {
		t.Fatalf("Content-Type = %q, want text/csv", got)
	}
	assertContains(t, rec.Header().Get("Content-Disposition"), "bill-of-fare-invoices.csv")
	assertContains(t, rec.Body.String(), "invoice_id,created_at,item_name,quantity,unit_price,subtotal,invoice_total", "Coffee,2,50,100,100")
}

func TestCreateInvoiceNonHTMXAndViewInvoiceErrors(t *testing.T) {
	h := newTestHandler(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/invoice/create", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "empty"})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusSeeOther)
	if rec.Header().Get("Location") != "/" {
		t.Fatalf("empty invoice redirect = %q, want /", rec.Header().Get("Location"))
	}

	session := "invoice-session"
	h.Cart.Add(session, models.MenuItem{ID: 7, Name: "Tea", Price: 20})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/invoice/create", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: session})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusSeeOther)
	if got := rec.Header().Get("Location"); got != "/invoice?id=1&print=1" {
		t.Fatalf("invoice redirect = %q, want /invoice?id=1&print=1", got)
	}

	rec = httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/invoice?id=404", nil))
	assertStatus(t, rec, http.StatusNotFound)
	assertContains(t, rec.Body.String(), "invoice not found")
}

func TestAdminItemMutationValidationErrors(t *testing.T) {
	h := newTestHandler(t)

	rec := httptest.NewRecorder()
	req := formRequest(http.MethodPost, "/admin/items/create", url.Values{"category_id": {"0"}, "name": {""}, "price": {"10"}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusBadRequest)
	assertContains(t, rec.Body.String(), "name and non-negative price are required")

	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/admin/items/update", url.Values{"id": {"0"}, "category_id": {"0"}, "name": {""}, "price": {"10"}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusBadRequest)
	assertContains(t, rec.Body.String(), "valid id, category, name")

	rec = httptest.NewRecorder()
	req = formRequest(http.MethodPost, "/admin/items/delete", url.Values{"id": {"0"}})
	h.Routes().ServeHTTP(rec, req)
	assertStatus(t, rec, http.StatusBadRequest)
	assertContains(t, rec.Body.String(), "valid id is required")
}

func TestHandlerHelpersDefaultValues(t *testing.T) {
	h := Handler{}
	if got := h.appVersion(); got != "dev" {
		t.Fatalf("appVersion empty = %q, want dev", got)
	}
	if got := menuPriceLabel(nil); got != "₹0" {
		t.Fatalf("menuPriceLabel nil = %q, want ₹0", got)
	}
	if got := menuColorClass(6); got != "category-color-1" {
		t.Fatalf("menuColorClass = %q, want category-color-1", got)
	}
}

func TestCreateInvoiceHTMXRendersPreviewAndClearsCart(t *testing.T) {
	h := newTestHandler(t)
	session := "session-1"
	h.Cart.Add(session, models.MenuItem{ID: 10, Name: "Chicken Tikka Kabab", Variant: "5pc", Price: 160})
	h.Cart.Add(session, models.MenuItem{ID: 20, Name: "Naan", Price: 30})

	req := httptest.NewRequest(http.MethodPost, "/invoice/create", nil)
	req.Header.Set("HX-Request", "true")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: session})
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{"Invoice Preview", "Order #1", "Chicken Tikka Kabab (5pc)", "₹190", `hx-swap-oob="innerHTML:#cart"`, "No items added yet"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q:\n%s", want, body)
		}
	}

	items, total := h.Cart.Snapshot(session)
	if len(items) != 0 || total != 0 {
		t.Fatalf("cart after invoice = len %d total %d, want empty zero", len(items), total)
	}
	sales := h.Cart.SessionSales(session)
	if sales.Count != 1 || sales.Total != 190 {
		t.Fatalf("sales = %+v, want count 1 total 190", sales)
	}
}

func TestCreateInvoiceHTMXWithEmptyCartReturnsNoContent(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/invoice/create", nil)
	req.Header.Set("HX-Request", "true")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "empty-session"})
	rec := httptest.NewRecorder()

	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", rec.Body.String())
	}
}

func TestViewInvoiceRendersRestaurantNameAndPrintScript(t *testing.T) {
	h := newTestHandler(t)
	if err := h.Settings.UpdateRestaurantName("Cafe Example"); err != nil {
		t.Fatalf("UpdateRestaurantName: %v", err)
	}
	id, err := h.Invoices.Create([]models.CartItem{
		{MenuItem: models.MenuItem{ID: 1, Name: "Soup", Price: 80}, Quantity: 1, Subtotal: 80},
	}, 80)
	if err != nil {
		t.Fatalf("Create invoice: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/invoice?id=1&print=1", nil)
	rec := httptest.NewRecorder()
	h.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Cafe Example", "Order #1", "Soup", "window.print()"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q:\n%s", want, body)
		}
	}
	if id != 1 {
		t.Fatalf("invoice id = %d, want 1", id)
	}
}
