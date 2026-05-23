package services

import "testing"

func TestSettingsServiceRestaurantName(t *testing.T) {
	database := openTestDB(t)
	settings := SettingsService{DB: database}

	if got := settings.RestaurantName(); got != "Bill of Fare" {
		t.Fatalf("default RestaurantName = %q, want Bill of Fare", got)
	}

	if err := settings.UpdateRestaurantName("  Cafe Test  "); err != nil {
		t.Fatalf("UpdateRestaurantName: %v", err)
	}
	if got := settings.RestaurantName(); got != "Cafe Test" {
		t.Fatalf("updated RestaurantName = %q, want Cafe Test", got)
	}
	if _, err := database.Exec("UPDATE app_settings SET value = '' WHERE key = 'restaurant_name'"); err != nil {
		t.Fatalf("blank restaurant setting: %v", err)
	}
	if got := settings.RestaurantName(); got != "Bill of Fare" {
		t.Fatalf("blank RestaurantName = %q, want fallback Bill of Fare", got)
	}
	if _, err := database.Exec("DELETE FROM app_settings WHERE key = 'restaurant_name'"); err != nil {
		t.Fatalf("delete restaurant setting: %v", err)
	}
	if got := settings.RestaurantName(); got != "Bill of Fare" {
		t.Fatalf("missing RestaurantName = %q, want fallback Bill of Fare", got)
	}

	if err := settings.UpdateRestaurantName("   "); err == nil {
		t.Fatal("UpdateRestaurantName blank succeeded, want error")
	}
}
