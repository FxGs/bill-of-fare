package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"bill-of-fare/internal/services"
)

type Handler struct {
	Tpl      *template.Template
	Menu     services.MenuService
	Cart     *services.CartService
	Invoices services.InvoiceService
	Static   http.Handler
}

func (h Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", h.Static))
	mux.HandleFunc("/", h.index)
	mux.HandleFunc("/menu/options", h.menuOptions)
	mux.HandleFunc("/cart/add", h.addToCart)
	mux.HandleFunc("/cart/qty", h.changeQty)
	mux.HandleFunc("/cart/remove", h.remove)
	mux.HandleFunc("/cart", h.cartFragment)
	mux.HandleFunc("/invoice/create", h.createInvoice)
	mux.HandleFunc("/invoice", h.viewInvoice)
	return mux
}

func (h Handler) index(w http.ResponseWriter, r *http.Request) {
	cats, _ := h.Menu.ListCategoriesWithItems()
	s := h.Cart.SessionID(w, r)
	items, total := h.Cart.Snapshot(s)
	_ = h.Tpl.ExecuteTemplate(w, "layout", map[string]any{"Categories": cats, "CartItems": items, "Total": total})
}

func (h Handler) menuOptions(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	catID, _ := strconv.Atoi(r.FormValue("category_id"))
	name := r.FormValue("name")
	group, err := h.Menu.GetMenuGroup(catID, name)
	if err != nil {
		http.Error(w, "menu item not found", http.StatusNotFound)
		return
	}
	_ = h.Tpl.ExecuteTemplate(w, "menu-options", map[string]any{"Group": group})
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
	items, total := h.Cart.Snapshot(h.Cart.SessionID(w, r))
	_ = h.Tpl.ExecuteTemplate(w, "cart", map[string]any{"CartItems": items, "Total": total})
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
	h.Cart.Clear(s)
	http.Redirect(w, r, "/invoice?id="+strconv.Itoa(id), http.StatusSeeOther)
}
func (h Handler) viewInvoice(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	inv, err := h.Invoices.Get(id)
	if err != nil {
		http.Error(w, "invoice not found", 404)
		return
	}
	_ = h.Tpl.ExecuteTemplate(w, "invoice", map[string]any{"Invoice": inv})
}
func atoi(s string) int { i, _ := strconv.Atoi(s); return i }
