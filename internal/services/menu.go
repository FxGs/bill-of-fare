package services

import (
	"database/sql"
	"fmt"

	"bill-of-fare/internal/models"
)

type MenuService struct{ DB *sql.DB }

func (s MenuService) ListCategoriesWithItems() ([]models.Category, error) {
	rows, err := s.DB.Query(`
		SELECT c.id, c.name, m.id, m.name, COALESCE(m.variant, ''), m.price
		FROM categories c
		JOIN menu_items m ON m.category_id = c.id
		ORDER BY c.name, m.name, m.variant`)
	if err != nil {
		return nil, fmt.Errorf("query menu: %w", err)
	}
	defer rows.Close()

	cats := []models.Category{}
	catIdx := map[int]int{}
	groupIdx := map[int]map[string]int{}

	for rows.Next() {
		var cID, mID, price int
		var cName, mName, variant string
		if err := rows.Scan(&cID, &cName, &mID, &mName, &variant, &price); err != nil {
			return nil, fmt.Errorf("scan menu row: %w", err)
		}
		if _, ok := catIdx[cID]; !ok {
			catIdx[cID] = len(cats)
			cats = append(cats, models.Category{ID: cID, Name: cName})
			groupIdx[cID] = map[string]int{}
		}
		cPos := catIdx[cID]
		if _, ok := groupIdx[cID][mName]; !ok {
			groupIdx[cID][mName] = len(cats[cPos].Menu)
			cats[cPos].Menu = append(cats[cPos].Menu, models.MenuGroup{Name: mName})
		}
		gPos := groupIdx[cID][mName]
		cats[cPos].Menu[gPos].Variants = append(cats[cPos].Menu[gPos].Variants, models.VariantOption{ID: mID, Label: variant, Price: price, HasName: variant != ""})
	}
	return cats, rows.Err()
}

func (s MenuService) GetMenuItem(id int) (models.MenuItem, error) {
	var m models.MenuItem
	err := s.DB.QueryRow(`
		SELECT m.id, m.category_id, c.name, m.name, COALESCE(m.variant, ''), m.price
		FROM menu_items m JOIN categories c ON c.id = m.category_id
		WHERE m.id = ?`, id).Scan(&m.ID, &m.CategoryID, &m.CategoryName, &m.Name, &m.Variant, &m.Price)
	if err != nil {
		return models.MenuItem{}, err
	}
	return m, nil
}

func (s MenuService) GetMenuGroup(categoryID int, itemName string) (models.MenuGroup, error) {
	rows, err := s.DB.Query(`
		SELECT id, COALESCE(variant, ''), price
		FROM menu_items
		WHERE category_id = ? AND name = ?
		ORDER BY variant`, categoryID, itemName)
	if err != nil {
		return models.MenuGroup{}, err
	}
	defer rows.Close()
	group := models.MenuGroup{Name: itemName}
	for rows.Next() {
		var id, price int
		var variant string
		if err := rows.Scan(&id, &variant, &price); err != nil {
			return models.MenuGroup{}, err
		}
		group.Variants = append(group.Variants, models.VariantOption{ID: id, Label: variant, Price: price, HasName: variant != ""})
	}
	if len(group.Variants) == 0 {
		return models.MenuGroup{}, sql.ErrNoRows
	}
	return group, nil
}
