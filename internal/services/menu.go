package services

import (
	"database/sql"
	"fmt"

	"bill-of-fare/internal/models"
)

type MenuService struct{ DB *sql.DB }

const BestSellersCategoryID = -1

func (s MenuService) ListCategoriesWithItems() ([]models.Category, error) {
	return s.listCategoriesWithItems(`
		SELECT c.id, c.name, m.id, m.name, COALESCE(m.variant, ''), m.price, m.available, m.best_seller
		FROM categories c
		JOIN menu_items m ON m.category_id = c.id
		ORDER BY c.name, m.name, m.variant`)
}

func (s MenuService) ListBestSellerItems() ([]models.Category, error) {
	return s.listCategoriesWithItems(`
		SELECT c.id, c.name, m.id, m.name, COALESCE(m.variant, ''), m.price, m.available, m.best_seller
		FROM categories c
		JOIN menu_items m ON m.category_id = c.id
		WHERE m.best_seller = 1
		ORDER BY c.name, m.name, m.variant`)
}

func (s MenuService) listCategoriesWithItems(query string) ([]models.Category, error) {
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query menu: %w", err)
	}
	defer rows.Close()

	cats := []models.Category{}
	idx := map[int]int{}

	for rows.Next() {
		var cID, mID, price int
		var available, bestSeller bool
		var cName, mName, variant string
		if err := rows.Scan(&cID, &cName, &mID, &mName, &variant, &price, &available, &bestSeller); err != nil {
			return nil, fmt.Errorf("scan menu row: %w", err)
		}
		if _, ok := idx[cID]; !ok {
			idx[cID] = len(cats)
			cats = append(cats, models.Category{ID: cID, Name: cName})
		}
		cats[idx[cID]].Items = append(cats[idx[cID]].Items, models.MenuItem{ID: mID, CategoryID: cID, CategoryName: cName, Name: mName, Variant: variant, Price: price, Available: available, BestSeller: bestSeller})
	}
	return cats, rows.Err()
}

func (s MenuService) GetMenuItem(id int) (models.MenuItem, error) {
	var m models.MenuItem
	err := s.DB.QueryRow(`
		SELECT m.id, m.category_id, c.name, m.name, COALESCE(m.variant, ''), m.price, m.available, m.best_seller
		FROM menu_items m JOIN categories c ON c.id = m.category_id
		WHERE m.id = ? AND m.available = 1`, id).Scan(&m.ID, &m.CategoryID, &m.CategoryName, &m.Name, &m.Variant, &m.Price, &m.Available, &m.BestSeller)
	if err != nil {
		return models.MenuItem{}, err
	}
	return m, nil
}

func (s MenuService) ListCategories() ([]models.Category, error) {
	rows, err := s.DB.Query("SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	cats := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (s MenuService) ListMenuItems() ([]models.MenuItem, error) {
	rows, err := s.DB.Query(`
		SELECT m.id, m.category_id, c.name, m.name, COALESCE(m.variant, ''), m.price, m.available, m.best_seller
		FROM menu_items m
		JOIN categories c ON c.id = m.category_id
		ORDER BY c.name, m.name, m.variant`)
	if err != nil {
		return nil, fmt.Errorf("query menu items: %w", err)
	}
	defer rows.Close()

	items := []models.MenuItem{}
	for rows.Next() {
		var m models.MenuItem
		if err := rows.Scan(&m.ID, &m.CategoryID, &m.CategoryName, &m.Name, &m.Variant, &m.Price, &m.Available, &m.BestSeller); err != nil {
			return nil, fmt.Errorf("scan menu item: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (s MenuService) CreateMenuItem(categoryID int, newCategory, name, variant string, price int) error {
	if name == "" || price < 0 {
		return fmt.Errorf("name and non-negative price are required")
	}
	cid, err := s.resolveCategoryID(categoryID, newCategory)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`INSERT INTO menu_items(category_id, name, variant, price, available, best_seller) VALUES(?,?,?,?,1,0)`, cid, name, variant, price)
	if err != nil {
		return fmt.Errorf("insert menu item: %w", err)
	}
	return nil
}

func (s MenuService) CreateCategory(name string) error {
	if name == "" {
		return fmt.Errorf("category name is required")
	}
	_, err := s.DB.Exec("INSERT INTO categories(name) VALUES(?)", name)
	if err != nil {
		return fmt.Errorf("insert category: %w", err)
	}
	return nil
}

func (s MenuService) DeleteCategory(id int) error {
	if id <= 0 {
		return fmt.Errorf("valid category is required")
	}
	var itemCount int
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM menu_items WHERE category_id = ?", id).Scan(&itemCount); err != nil {
		return fmt.Errorf("check category items: %w", err)
	}
	if itemCount > 0 {
		return fmt.Errorf("delete or move %d menu items before deleting this category", itemCount)
	}
	_, err := s.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

func (s MenuService) UpdateMenuItem(id, categoryID int, name, variant string, price int, available, bestSeller bool) error {
	if id <= 0 || categoryID <= 0 || name == "" || price < 0 {
		return fmt.Errorf("valid id, category, name, and non-negative price are required")
	}
	_, err := s.DB.Exec(`UPDATE menu_items SET category_id = ?, name = ?, variant = ?, price = ?, available = ?, best_seller = ? WHERE id = ?`, categoryID, name, variant, price, available, bestSeller, id)
	if err != nil {
		return fmt.Errorf("update menu item: %w", err)
	}
	return nil
}

func (s MenuService) DeleteMenuItem(id int) error {
	if id <= 0 {
		return fmt.Errorf("valid id is required")
	}
	_, err := s.DB.Exec("DELETE FROM menu_items WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete menu item: %w", err)
	}
	return nil
}

func (s MenuService) resolveCategoryID(categoryID int, newCategory string) (int, error) {
	if newCategory != "" {
		if _, err := s.DB.Exec("INSERT OR IGNORE INTO categories(name) VALUES(?)", newCategory); err != nil {
			return 0, fmt.Errorf("insert category: %w", err)
		}
		var cid int
		if err := s.DB.QueryRow("SELECT id FROM categories WHERE name = ?", newCategory).Scan(&cid); err != nil {
			return 0, fmt.Errorf("lookup category: %w", err)
		}
		return cid, nil
	}
	if categoryID <= 0 {
		return 0, fmt.Errorf("category is required")
	}
	return categoryID, nil
}
