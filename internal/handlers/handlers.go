package handlers

import (
	"encoding/csv"
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"bill-of-fare/internal/models"
	"bill-of-fare/internal/services"
)

type Handler struct {
	Tpl      *template.Template
	Menu     services.MenuService
	Cart     *services.CartService
	Invoices services.InvoiceService
	Settings services.SettingsService
	Version  string
	Static   http.Handler
}

type menuDisplayCategory struct {
	ID         int
	Name       string
	ColorClass string
	Items      []menuDisplayItem
}

type menuDisplayItem struct {
	ID            int
	CategoryName  string
	Name          string
	Variants      []models.MenuItem
	PriceLabel    string
	VariantLabel  string
	HasChoices    bool
	Available     bool
	BestSeller    bool
	SelectedCount int
	CountKey      string
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
	mux.HandleFunc("/cart/toggle", h.toggleCart)
	mux.HandleFunc("/cart/qty", h.changeQty)
	mux.HandleFunc("/cart/remove", h.remove)
	mux.HandleFunc("/cart/clear", h.clearCart)
	mux.HandleFunc("/cart", h.cartFragment)
	mux.HandleFunc("/sales", h.sales)
	mux.HandleFunc("/sales/invoices/export", h.exportInvoices)
	mux.HandleFunc("/sales/invoices/preview", h.invoicePreview)
	mux.HandleFunc("/admin", h.admin)
	mux.HandleFunc("/admin/categories/create", h.adminCreateCategory)
	mux.HandleFunc("/admin/categories/delete", h.adminDeleteCategory)
	mux.HandleFunc("/admin/items/create", h.adminCreateItem)
	mux.HandleFunc("/admin/items/update", h.adminUpdateItem)
	mux.HandleFunc("/admin/items/delete", h.adminDeleteItem)
	mux.HandleFunc("/admin/invoices/export", h.exportInvoices)
	mux.HandleFunc("/admin/settings/restaurant-name", h.adminUpdateRestaurantName)
	mux.HandleFunc("/invoice/create", h.createInvoice)
	return mux
}

func (h Handler) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := h.pageData(w, r, 0)
	if invoiceID := atoi(r.URL.Query().Get("invoice_id")); invoiceID > 0 {
		if inv, err := h.Invoices.Get(invoiceID); err == nil {
			data["Invoice"] = inv
			data["AutoOpenInvoicePreview"] = true
		}
	}
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
func (h Handler) toggleCart(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id, _ := strconv.Atoi(r.FormValue("item_id"))
	if it, err := h.Menu.GetMenuItem(id); err == nil {
		h.Cart.Toggle(h.Cart.SessionID(w, r), it)
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
func (h Handler) clearCart(w http.ResponseWriter, r *http.Request) {
	h.Cart.Clear(h.Cart.SessionID(w, r))
	h.cartFragment(w, r)
}
func (h Handler) cartFragment(w http.ResponseWriter, r *http.Request) {
	s := h.Cart.SessionID(w, r)
	items, total := h.Cart.Snapshot(s)
	orderNumber, _ := h.Invoices.NextNumber()
	_ = h.Tpl.ExecuteTemplate(w, "cart", h.cartData(items, total, orderNumber))
}
func (h Handler) sales(w http.ResponseWriter, r *http.Request) {
	s := h.Cart.SessionID(w, r)
	_, currentTotal := h.Cart.Snapshot(s)
	today, _ := h.Invoices.TodaySales()
	sessionSales := h.Cart.SessionSales(s)
	invoices, _ := h.Invoices.List(100)
	average := 0
	if sessionSales.Count > 0 {
		average = sessionSales.Total / sessionSales.Count
	}
	_ = h.Tpl.ExecuteTemplate(w, "sales", map[string]any{
		"ActivePage":        "sales",
		"AppVersion":        h.appVersion(),
		"AverageOrderValue": average,
		"CurrentTotal":      currentTotal,
		"Invoices":          invoices,
		"RestaurantName":    h.Settings.RestaurantName(),
		"SessionSales":      sessionSales,
		"TodaySales":        today,
	})
}
func (h Handler) admin(w http.ResponseWriter, r *http.Request) {
	cats, _ := h.Menu.ListCategories()
	items, _ := h.Menu.ListMenuItems()
	_ = h.Tpl.ExecuteTemplate(w, "admin", map[string]any{"ActivePage": "admin", "AppVersion": h.appVersion(), "Categories": cats, "Items": items, "RestaurantName": h.Settings.RestaurantName()})
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
	err := h.Menu.UpdateMenuItem(
		atoi(r.FormValue("id")),
		atoi(r.FormValue("category_id")),
		r.FormValue("name"),
		r.FormValue("variant"),
		atoi(r.FormValue("price")),
		r.FormValue("available") == "on",
		r.FormValue("best_seller") == "on",
	)
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
func (h Handler) exportInvoices(w http.ResponseWriter, r *http.Request) {
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
		if r.Header.Get("HX-Request") == "true" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
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
	if r.Header.Get("HX-Request") == "true" {
		inv, err := h.Invoices.Get(id)
		if err != nil {
			http.Error(w, "invoice not found", 404)
			return
		}
		orderNumber, _ := h.Invoices.NextNumber()
		_ = h.Tpl.ExecuteTemplate(w, "invoice-created", map[string]any{
			"Invoice":        inv,
			"RestaurantName": h.Settings.RestaurantName(),
			"CartData":       h.cartData([]models.CartItem{}, 0, orderNumber),
		})
		return
	}
	http.Redirect(w, r, "/?invoice_id="+strconv.Itoa(id), http.StatusSeeOther)
}

func (h Handler) invoicePreview(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	inv, err := h.Invoices.Get(id)
	if err != nil {
		http.Error(w, "invoice not found", 404)
		return
	}
	_ = h.Tpl.ExecuteTemplate(w, "invoice-preview-fragment", map[string]any{"Invoice": inv, "RestaurantName": h.Settings.RestaurantName()})
}
func atoi(s string) int { i, _ := strconv.Atoi(s); return i }

func (h Handler) pageData(w http.ResponseWriter, r *http.Request, selectedID int) map[string]any {
	cats, _ := h.Menu.ListCategoriesWithItems()
	bestSellerCats, _ := h.Menu.ListBestSellerItems()
	menuCats := cats
	colorByCategoryID := map[int]string{}
	categoryTabs := make([]menuCategoryTab, 0, len(cats)+1)
	if len(bestSellerCats) > 0 {
		bestSellerColor := menuColorClass(0)
		colorByCategoryID[services.BestSellersCategoryID] = bestSellerColor
		categoryTabs = append(categoryTabs, menuCategoryTab{ID: services.BestSellersCategoryID, Name: "Best Sellers", ColorClass: bestSellerColor})
	}
	for i, c := range cats {
		color := menuColorClass(i + 1)
		colorByCategoryID[c.ID] = color
		categoryTabs = append(categoryTabs, menuCategoryTab{ID: c.ID, Name: c.Name, ColorClass: color})
	}
	if selectedID == services.BestSellersCategoryID {
		menuCats = bestSellerPreset(bestSellerCats)
	} else if selectedID > 0 {
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
	selectedCounts := menuSelectedCounts(items)
	itemCount := menuItemCount(items)
	orderNumber, _ := h.Invoices.NextNumber()
	return map[string]any{"ActivePage": "pos", "AppVersion": h.appVersion(), "Categories": cats, "CategoryTabs": categoryTabs, "MenuCategories": groupMenuCategories(menuCats, colorByCategoryID, selectedCounts), "SelectedCategoryID": selectedID, "CartItems": items, "ItemCount": itemCount, "Total": total, "OrderNumber": orderNumber, "PreviewTime": time.Now(), "RestaurantName": h.Settings.RestaurantName()}
}

func (h Handler) cartData(items []models.CartItem, total, orderNumber int) map[string]any {
	counts := menuSelectedCounts(items)
	countsJSON, _ := json.Marshal(counts)
	return map[string]any{
		"CartItems":      items,
		"ItemCount":      menuItemCount(items),
		"Total":          total,
		"OrderNumber":    orderNumber,
		"PreviewTime":    time.Now(),
		"RestaurantName": h.Settings.RestaurantName(),
		"MenuCountsJSON": template.JS(countsJSON),
	}
}

func menuItemCount(items []models.CartItem) int {
	count := 0
	for _, item := range items {
		count += item.Quantity
	}
	return count
}

func (h Handler) appVersion() string {
	if h.Version == "" {
		return "dev"
	}
	return h.Version
}

func groupMenuCategories(cats []models.Category, colorByCategoryID map[int]string, selectedCounts map[string]int) []menuDisplayCategory {
	groupedCats := make([]menuDisplayCategory, 0, len(cats))
	for _, c := range cats {
		displayCat := menuDisplayCategory{ID: c.ID, Name: c.Name, ColorClass: colorByCategoryID[c.ID]}
		itemIndex := map[string]int{}
		for _, item := range c.Items {
			groupKey := menuCountKey(item.CategoryID, item.Name)
			if _, ok := itemIndex[groupKey]; !ok {
				itemIndex[groupKey] = len(displayCat.Items)
				displayCat.Items = append(displayCat.Items, menuDisplayItem{
					ID:           item.ID,
					CategoryName: item.CategoryName,
					Name:         item.Name,
				})
			}
			idx := itemIndex[groupKey]
			displayCat.Items[idx].Variants = append(displayCat.Items[idx].Variants, item)
		}
		for i := range displayCat.Items {
			sortMenuVariants(displayCat.Items[i].Variants)
			displayCat.Items[i].HasChoices = len(displayCat.Items[i].Variants) > 1
			displayCat.Items[i].Available = menuItemAvailable(displayCat.Items[i].Variants)
			displayCat.Items[i].BestSeller = menuItemBestSeller(displayCat.Items[i].Variants)
			displayCat.Items[i].PriceLabel = menuPriceLabel(displayCat.Items[i].Variants)
			displayCat.Items[i].VariantLabel = menuVariantLabel(displayCat.Items[i].Variants)
			displayCat.Items[i].CountKey = menuCountKey(displayCat.Items[i].Variants[0].CategoryID, displayCat.Items[i].Name)
			displayCat.Items[i].SelectedCount = selectedCounts[displayCat.Items[i].CountKey]
		}
		groupedCats = append(groupedCats, displayCat)
	}
	return groupedCats
}

func sortMenuVariants(items []models.MenuItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Price != items[j].Price {
			return items[i].Price < items[j].Price
		}
		if items[i].Variant != items[j].Variant {
			return items[i].Variant < items[j].Variant
		}
		return items[i].ID < items[j].ID
	})
}

func bestSellerPreset(cats []models.Category) []models.Category {
	items := []models.MenuItem{}
	for _, c := range cats {
		items = append(items, c.Items...)
	}
	if len(items) == 0 {
		return nil
	}
	return []models.Category{{ID: services.BestSellersCategoryID, Name: "Best Sellers", Items: items}}
}

func menuItemAvailable(items []models.MenuItem) bool {
	for _, item := range items {
		if item.Available {
			return true
		}
	}
	return false
}

func menuItemBestSeller(items []models.MenuItem) bool {
	for _, item := range items {
		if item.BestSeller {
			return true
		}
	}
	return false
}

func menuSelectedCounts(items []models.CartItem) map[string]int {
	counts := map[string]int{}
	for _, item := range items {
		counts[menuCountKey(item.MenuItem.CategoryID, item.MenuItem.Name)] += item.Quantity
	}
	return counts
}

func menuCountKey(categoryID int, name string) string {
	return strconv.Itoa(categoryID) + ":" + name
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
	if len(items) > 1 {
		return "from ₹" + strconv.Itoa(min)
	}
	if min == max {
		return "₹" + strconv.Itoa(min)
	}
	return "from ₹" + strconv.Itoa(min)
}

func menuVariantLabel(items []models.MenuItem) string {
	if len(items) == 0 {
		return ""
	}
	labels := make([]string, 0, len(items))
	for _, item := range items {
		if item.Variant == "" {
			labels = append(labels, "Regular")
			continue
		}
		labels = append(labels, item.Variant)
	}
	return strings.Join(labels, " / ")
}
