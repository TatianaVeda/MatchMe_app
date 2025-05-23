package controllers

import (
	"encoding/json"
	"net/http"

	"m/backend/models"

	"gorm.io/gorm"
)

var finnishCities = []City{
	{"Helsinki", 60.1699, 24.9384},
	{"Espoo", 60.2055, 24.6559},
	{"Vantaa", 60.2934, 25.0378},
	{"Turku", 60.4518, 22.2666},
	{"Tampere", 61.4981, 23.7610},
	{"Oulu", 65.0121, 25.4651},
	{"Lahti", 60.9827, 25.6615},
	{"Kuopio", 62.8924, 27.6770},
	{"Pori", 61.4850, 21.7973},
	{"Jyväskylä", 62.2426, 25.7473},
}

type City struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var citiesDB *gorm.DB

func InitCitiesController(db *gorm.DB) {
	citiesDB = db
}

func GetCities(w http.ResponseWriter, r *http.Request) {
	type row struct {
		City      string
		Latitude  float64
		Longitude float64
	}
	var rows []row
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
