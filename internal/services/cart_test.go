package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bill-of-fare/internal/models"
)

func TestCartAddChangeRemoveSnapshotAndSales(t *testing.T) {
	cart := NewCartService()
	session := "session-1"
	item := models.MenuItem{ID: 7, Name: "Paneer Tikka", Price: 180}

	cart.Add(session, item)
	cart.Add(session, item)

	items, total := cart.Snapshot(session)
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if items[0].Quantity != 2 || items[0].Subtotal != 360 || total != 360 {
		t.Fatalf("snapshot = %+v total %d, want qty 2 subtotal 360 total 360", items[0], total)
	}

	cart.ChangeQty(session, "7", -1)
	items, total = cart.Snapshot(session)
	if items[0].Quantity != 1 || total != 180 {
		t.Fatalf("after decrement = %+v total %d, want qty 1 total 180", items[0], total)
	}
	cart.ChangeQty(session, "missing", 1)
	items, total = cart.Snapshot(session)
	if len(items) != 1 || items[0].Quantity != 1 || total != 180 {
		t.Fatalf("after missing change = %+v total %d, want unchanged", items, total)
	}

	cart.RecordSale(session, total)
	sales := cart.SessionSales(session)
	if sales.Count != 1 || sales.Total != 180 {
		t.Fatalf("sales = %+v, want count 1 total 180", sales)
	}

	cart.ChangeQty(session, "7", -1)
	items, total = cart.Snapshot(session)
	if len(items) != 0 || total != 0 {
		t.Fatalf("after zero qty len = %d total = %d, want empty zero", len(items), total)
	}

	cart.Add(session, item)
	cart.Remove(session, "7")
	items, total = cart.Snapshot(session)
	if len(items) != 0 || total != 0 {
		t.Fatalf("after remove len = %d total = %d, want empty zero", len(items), total)
	}

	cart.Add(session, item)
	cart.Clear(session)
	items, total = cart.Snapshot(session)
	if len(items) != 0 || total != 0 {
		t.Fatalf("after clear len = %d total = %d, want empty zero", len(items), total)
	}
}

func TestSessionIDReusesExistingCookieOrSetsNewOne(t *testing.T) {
	cart := NewCartService()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "existing"})
	rec := httptest.NewRecorder()
	if got := cart.SessionID(rec, req); got != "existing" {
		t.Fatalf("SessionID existing = %q, want existing", got)
	}
	if cookies := rec.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("unexpected cookies set: %+v", cookies)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	got := cart.SessionID(rec, req)
	if got == "" {
		t.Fatal("SessionID new returned empty id")
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session_id" || cookies[0].Value != got {
		t.Fatalf("set cookies = %+v, want session_id cookie with generated id %q", cookies, got)
	}
}
