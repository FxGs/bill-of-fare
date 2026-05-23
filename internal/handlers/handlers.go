package handlers

import (
	"encoding/csv"
	"html/template"
	"net/http"
	"strconv"

	"bill-of-fare/internal/models"
	"bill-of-fare/internal/services"
)

type Handler struct {
	Tpl      *template.Template
	Menu     services.MenuService
	Cart     *services.CartService
	Invoices services.InvoiceService
	Settings services.SettingsService
	Static   http.Handler
}

type menuDisplayCategory struct {
	ID         int
	Name       string
	ColorClass string
	Items      []menuDisplayItem
}

type menuDisplayItem struct {
	ID           int
	CategoryName string
	Name         string
	Variants     []models.MenuItem
	PriceLabel   string
	HasChoices   bool
}

type menuCategoryTab struct {
	ID         int
	Name       string
	ColorClass string
}

func (h Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", h.Static))
	mux.HandleFunc("/", h.index)
	mux.HandleFunc("/menu", h.menuPane)
	mux.HandleFunc("/cart/add", h.addToCart)
	mux.HandleFunc("/cart/qty", h.changeQty)
	mux.HandleFunc("/cart/remove", h.remove)
	mux.HandleFunc("/cart", h.cartFragment)
	mux.HandleFunc("/sales", h.salesSummary)
	mux.HandleFunc("/admin", h.admin)
	mux.HandleFunc("/admin/categories/create", h.adminCreateCategory)
	mux.HandleFunc("/admin/categories/delete", h.adminDeleteCategory)
	mux.HandleFunc("/admin/items/create", h.adminCreateItem)
	mux.HandleFunc("/admin/items/update", h.adminUpdateItem)
	mux.HandleFunc("/admin/items/delete", h.adminDeleteItem)
	mux.HandleFunc("/admin/invoices/export", h.adminExportInvoices)
	mux.HandleFunc("/admin/settings/restaurant-name", h.adminUpdateRestaurantName)
	mux.HandleFunc("/invoice/create", h.createInvoice)
	mux.HandleFunc("/invoice", h.viewInvoice)
	return mux
}

func (h Handler) index(w http.ResponseWriter, r *http.Request) {
	data := h.pageData(w, r, 0)
	_ = h.Tpl.ExecuteTemplate(w, "layout", data)
}

func (h Handler) menuPane(w http.ResponseWriter, r *http.Request) {
	selectedID := atoi(r.URL.Query().Get("category_id"))
	data := h.pageData(w, r, selectedID)
	_ = h.Tpl.ExecuteTemplate(w, "menu-pane", data)
}
func (h Handler) addToCart(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("item_id"))
	if it, err := h.Menu.GetMenuItem(id); err == nil {
		h.Cart.Add(h.Cart.SessionID(w, r), it)
	}
	h.cartFragment(w, r)
}
func (h Handler) changeQty(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	h.Cart.ChangeQty(h.Cart.SessionID(w, r), r.FormValue("key"), atoi(r.FormValue("delta")))
	h.cartFragment(w, r)
}
func (h Handler) remove(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	h.Cart.Remove(h.Cart.SessionID(w, r), r.FormValue("key"))
	h.cartFragment(w, r)
}
func (h Handler) cartFragment(w http.ResponseWriter, r *http.Request) {
	s := h.Cart.SessionID(w, r)
	items, total := h.Cart.Snapshot(s)
	orderNumber, _ := h.Invoices.NextNumber()
	_ = h.Tpl.ExecuteTemplate(w, "cart", map[string]any{"CartItems": items, "Total": total, "OrderNumber": orderNumber})
}
func (h Handler) salesSummary(w http.ResponseWriter, r *http.Request) {
	s := h.Cart.SessionID(w, r)
	_, currentTotal := h.Cart.Snapshot(s)
	today, _ := h.Invoices.TodaySales()
	_ = h.Tpl.ExecuteTemplate(w, "sales-summary", map[string]any{"TodaySales": today, "SessionSales": h.Cart.SessionSales(s), "CurrentTotal": currentTotal})
}
func (h Handler) admin(w http.ResponseWriter, r *http.Request) {
	cats, _ := h.Menu.ListCategories()
	items, _ := h.Menu.ListMenuItems()
	invoices, _ := h.Invoices.List(100)
	_ = h.Tpl.ExecuteTemplate(w, "admin", map[string]any{"Categories": cats, "Items": items, "Invoices": invoices, "RestaurantName": h.Settings.RestaurantName()})
}
func (h Handler) adminCreateCategory(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if err := h.Menu.CreateCategory(r.FormValue("name")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func (h Handler) adminDeleteCategory(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if err := h.Menu.DeleteCategory(atoi(r.FormValue("id"))); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func (h Handler) adminCreateItem(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	err := h.Menu.CreateMenuItem(atoi(r.FormValue("category_id")), "", r.FormValue("name"), r.FormValue("variant"), atoi(r.FormValue("price")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func (h Handler) adminUpdateItem(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	err := h.Menu.UpdateMenuItem(atoi(r.FormValue("id")), atoi(r.FormValue("category_id")), r.FormValue("name"), r.FormValue("variant"), atoi(r.FormValue("price")))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func (h Handler) adminDeleteItem(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if err := h.Menu.DeleteMenuItem(atoi(r.FormValue("id"))); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func (h Handler) adminExportInvoices(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Invoices.ExportRows()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", `attachment; filename="bill-of-fare-invoices.csv"`)

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"invoice_id", "created_at", "item_name", "quantity", "unit_price", "subtotal", "invoice_total"})
	for _, row := range rows {
		_ = cw.Write([]string{
			strconv.Itoa(row.InvoiceID),
			row.CreatedAt.Format("2006-01-02 15:04:05"),
			row.ItemName,
			strconv.Itoa(row.Quantity),
			strconv.Itoa(row.UnitPrice),
			strconv.Itoa(row.Subtotal),
			strconv.Itoa(row.InvoiceTotal),
		})
	}
	cw.Flush()
}
func (h Handler) adminUpdateRestaurantName(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if err := h.Settings.UpdateRestaurantName(r.FormValue("restaurant_name")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func (h Handler) createInvoice(w http.ResponseWriter, r *http.Request) {
	s := h.Cart.SessionID(w, r)
	items, total := h.Cart.Snapshot(s)
	if len(items) == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id, err := h.Invoices.Create(items, total)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	h.Cart.RecordSale(s, total)
	h.Cart.Clear(s)
	http.Redirect(w, r, "/invoice?id="+strconv.Itoa(id)+"&print=1", http.StatusSeeOther)
}
func (h Handler) viewInvoice(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	inv, err := h.Invoices.Get(id)
	if err != nil {
		http.Error(w, "invoice not found", 404)
		return
	}
	_ = h.Tpl.ExecuteTemplate(w, "invoice", map[string]any{"Invoice": inv, "RestaurantName": h.Settings.RestaurantName(), "AutoPrint": r.URL.Query().Get("print") == "1"})
}
func atoi(s string) int { i, _ := strconv.Atoi(s); return i }

func (h Handler) pageData(w http.ResponseWriter, r *http.Request, selectedID int) map[string]any {
	cats, _ := h.Menu.ListCategoriesWithItems()
	menuCats := cats
	colorByCategoryID := map[int]string{}
	categoryTabs := make([]menuCategoryTab, 0, len(cats))
	for i, c := range cats {
		color := menuColorClass(i + 1)
		colorByCategoryID[c.ID] = color
		categoryTabs = append(categoryTabs, menuCategoryTab{ID: c.ID, Name: c.Name, ColorClass: color})
	}
	if selectedID > 0 {
		menuCats = nil
		for _, c := range cats {
			if c.ID == selectedID {
				menuCats = append(menuCats, c)
				break
			}
		}
	}
	s := h.Cart.SessionID(w, r)
	items, total := h.Cart.Snapshot(s)
	orderNumber, _ := h.Invoices.NextNumber()
	return map[string]any{"Categories": cats, "CategoryTabs": categoryTabs, "MenuCategories": groupMenuCategories(menuCats, colorByCategoryID), "SelectedCategoryID": selectedID, "CartItems": items, "Total": total, "OrderNumber": orderNumber}
}

func groupMenuCategories(cats []models.Category, colorByCategoryID map[int]string) []menuDisplayCategory {
	groupedCats := make([]menuDisplayCategory, 0, len(cats))
	for _, c := range cats {
		displayCat := menuDisplayCategory{ID: c.ID, Name: c.Name, ColorClass: colorByCategoryID[c.ID]}
		itemIndex := map[string]int{}
		for _, item := range c.Items {
			if _, ok := itemIndex[item.Name]; !ok {
				itemIndex[item.Name] = len(displayCat.Items)
				displayCat.Items = append(displayCat.Items, menuDisplayItem{
					ID:           item.ID,
					CategoryName: item.CategoryName,
					Name:         item.Name,
				})
			}
			idx := itemIndex[item.Name]
			displayCat.Items[idx].Variants = append(displayCat.Items[idx].Variants, item)
		}
		for i := range displayCat.Items {
			displayCat.Items[i].HasChoices = len(displayCat.Items[i].Variants) > 1
			displayCat.Items[i].PriceLabel = menuPriceLabel(displayCat.Items[i].Variants)
		}
		groupedCats = append(groupedCats, displayCat)
	}
	return groupedCats
}

func menuColorClass(index int) string {
	return "category-color-" + strconv.Itoa((index%6)+1)
}

func menuPriceLabel(items []models.MenuItem) string {
	if len(items) == 0 {
		return "₹0"
	}
	min := items[0].Price
	max := items[0].Price
	for _, item := range items[1:] {
		if item.Price < min {
			min = item.Price
		}
		if item.Price > max {
			max = item.Price
		}
	}
	if min == max {
		return "₹" + strconv.Itoa(min)
	}
	return "₹" + strconv.Itoa(min) + " - ₹" + strconv.Itoa(max)
}
