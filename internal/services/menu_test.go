package services

import "testing"

func TestMenuServiceCreatesUpdatesListsAndDeletesItems(t *testing.T) {
	database := openTestDB(t)
	menu := MenuService{DB: database}

	if err := menu.CreateCategory("Starters"); err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	cats, err := menu.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(cats) != 1 || cats[0].Name != "Starters" {
		t.Fatalf("categories = %+v, want Starters", cats)
	}

	if err := menu.CreateMenuItem(cats[0].ID, "", "Bruschetta", "Regular", 95); err != nil {
		t.Fatalf("CreateMenuItem: %v", err)
	}
	items, err := menu.ListMenuItems()
	if err != nil {
		t.Fatalf("ListMenuItems: %v", err)
	}
	if len(items) != 1 || items[0].Name != "Bruschetta" || items[0].Variant != "Regular" || items[0].Price != 95 {
		t.Fatalf("items = %+v, want Bruschetta Regular 95", items)
	}

	if err := menu.UpdateMenuItem(items[0].ID, cats[0].ID, "Tomato Bruschetta", "", 105); err != nil {
		t.Fatalf("UpdateMenuItem: %v", err)
	}
	updated, err := menu.GetMenuItem(items[0].ID)
	if err != nil {
		t.Fatalf("GetMenuItem: %v", err)
	}
	if updated.Name != "Tomato Bruschetta" || updated.Variant != "" || updated.Price != 105 {
		t.Fatalf("updated item = %+v, want Tomato Bruschetta blank variant 105", updated)
	}

	if err := menu.DeleteCategory(cats[0].ID); err == nil {
		t.Fatal("DeleteCategory with item succeeded, want error")
	}
	if err := menu.DeleteMenuItem(items[0].ID); err != nil {
		t.Fatalf("DeleteMenuItem: %v", err)
	}
	if err := menu.DeleteCategory(cats[0].ID); err != nil {
		t.Fatalf("DeleteCategory empty: %v", err)
	}
	cats, err = menu.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories after delete: %v", err)
	}
	if len(cats) != 0 {
		t.Fatalf("categories after delete = %+v, want empty", cats)
	}
}

func TestCreateMenuItemCanCreateNewCategory(t *testing.T) {
	database := openTestDB(t)
	menu := MenuService{DB: database}

	if err := menu.CreateMenuItem(0, "Desserts", "Panna Cotta", "", 110); err != nil {
		t.Fatalf("CreateMenuItem new category: %v", err)
	}
	items, err := menu.ListMenuItems()
	if err != nil {
		t.Fatalf("ListMenuItems: %v", err)
	}
	if len(items) != 1 || items[0].CategoryName != "Desserts" {
		t.Fatalf("items = %+v, want one Desserts item", items)
	}

	cats, err := menu.ListCategoriesWithItems()
	if err != nil {
		t.Fatalf("ListCategoriesWithItems: %v", err)
	}
	if len(cats) != 1 || cats[0].Name != "Desserts" || len(cats[0].Items) != 1 || cats[0].Items[0].Name != "Panna Cotta" {
		t.Fatalf("categories with items = %+v, want Desserts with Panna Cotta", cats)
	}
}

func TestMenuServiceValidation(t *testing.T) {
	database := openTestDB(t)
	menu := MenuService{DB: database}

	if err := menu.CreateCategory(""); err == nil {
		t.Fatal("CreateCategory empty succeeded, want error")
	}
	if err := menu.CreateMenuItem(0, "", "Soup", "", 100); err == nil {
		t.Fatal("CreateMenuItem without category succeeded, want error")
	}
	if err := menu.CreateMenuItem(1, "", "", "", 100); err == nil {
		t.Fatal("CreateMenuItem without name succeeded, want error")
	}
	if err := menu.CreateMenuItem(1, "", "Soup", "", -1); err == nil {
		t.Fatal("CreateMenuItem negative price succeeded, want error")
	}
	if err := menu.UpdateMenuItem(0, 1, "Soup", "", 100); err == nil {
		t.Fatal("UpdateMenuItem invalid id succeeded, want error")
	}
	if err := menu.DeleteMenuItem(0); err == nil {
		t.Fatal("DeleteMenuItem invalid id succeeded, want error")
	}
	if err := menu.DeleteCategory(0); err == nil {
		t.Fatal("DeleteCategory invalid id succeeded, want error")
	}
	if _, err := menu.GetMenuItem(999); err == nil {
		t.Fatal("GetMenuItem missing item succeeded, want error")
	}
}
