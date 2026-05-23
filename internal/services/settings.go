package services

import (
	"database/sql"
	"fmt"
	"strings"
)

const defaultRestaurantName = "Bill of Fare"

type SettingsService struct{ DB *sql.DB }

func (s SettingsService) RestaurantName() string {
	var name string
	err := s.DB.QueryRow("SELECT value FROM app_settings WHERE key = 'restaurant_name'").Scan(&name)
	if err != nil || strings.TrimSpace(name) == "" {
		return defaultRestaurantName
	}
	return name
}

func (s SettingsService) UpdateRestaurantName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("restaurant name is required")
	}
	_, err := s.DB.Exec(`
		INSERT INTO app_settings(key, value) VALUES('restaurant_name', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, name)
	if err != nil {
		return fmt.Errorf("update restaurant name: %w", err)
	}
	return nil
}
