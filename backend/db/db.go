package db

import (
	"database/sql"
	"m/backend/config"

	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("postgres", config.AppConfig.DatabaseURL)
	if err != nil {
		logrus.Fatalf("Error connecting to DB: %v", err)
	}
	if err = DB.Ping(); err != nil {
		logrus.Fatalf("Error checking connection to DB: %v", err)
	}
	logrus.Info("✅ Connected to DB")
}
