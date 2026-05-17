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
	idx := map[int]int{}

	for rows.Next() {
		var cID, mID, price int
		var cName, mName, variant string
		if err := rows.Scan(&cID, &cName, &mID, &mName, &variant, &price); err != nil {
			return nil, fmt.Errorf("scan menu row: %w", err)
		}
		if _, ok := idx[cID]; !ok {
			idx[cID] = len(cats)
			cats = append(cats, models.Category{ID: cID, Name: cName})
		}
		cats[idx[cID]].Items = append(cats[idx[cID]].Items, models.MenuItem{ID: mID, CategoryID: cID, CategoryName: cName, Name: mName, Variant: variant, Price: price})
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
