package controllers

import (
	"encoding/json"
	"m/backend/config"
	"net/http"
)

type City struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// GetCities возвращает список всех городов и их координат
func GetCities(w http.ResponseWriter, r *http.Request) {
	cities := make([]City, 0, len(config.AppConfig.CityCoords))
	for name, coords := range config.AppConfig.CityCoords {
		cities = append(cities, City{
			Name:      name,
			Latitude:  coords[0],
			Longitude: coords[1],
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cities)
}
