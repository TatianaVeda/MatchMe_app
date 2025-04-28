package controllers

import (
	"encoding/json"
	"net/http"

	"m/backend/models"

	"gorm.io/gorm"
)

type City struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var citiesDB *gorm.DB

func InitCitiesController(db *gorm.DB) {
	citiesDB = db
}

// GET /cities
func GetCities(w http.ResponseWriter, r *http.Request) {
	// 1) вытягиваем все непустые города (DISTINCT)
	type row struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	var rows []row
	// Берём первую встречающуюся пару (latitude, longitude) для каждого города
	err := citiesDB.
		Model(&models.Profile{}).
		Select("DISTINCT ON (city) city, latitude, longitude").
		Where("city <> ''").
		Order("city, id").
		Scan(&rows).Error
	if err != nil {
		http.Error(w, "Error fetching cities", http.StatusInternalServerError)
		return
	}

	// 2) мапим в выходную структуру
	cities := make([]City, len(rows))
	for i, r := range rows {
		cities[i] = City{
			Name:      r.City,
			Latitude:  r.Latitude,
			Longitude: r.Longitude,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cities)
}
