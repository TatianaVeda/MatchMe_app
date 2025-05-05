/* package main

import (
	"flag"
	"log"
	"m/backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	createAdmin := flag.Bool("createAdmin", false, "Create admin user")
	resetDB := flag.Bool("resetDB", false, "Reset database")
	flag.Parse()

	dsn := "host=localhost user=user password=password dbname=sopostavmenya port=5433 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if *resetDB {
		resetDatabase(db)
	}

	if *createAdmin {
		runAdminSetup(db)
	}
}

func resetDatabase(db *gorm.DB) {
	modelsToDrop := []interface{}{
		&models.User{}, &models.Profile{}, &models.Bio{}, &models.Preference{},
		&models.Recommendation{}, &models.Connection{}, &models.Chat{},
		&models.Message{}, &models.FakeUser{},
	}
	if err := db.Migrator().DropTable(modelsToDrop...); err != nil {
		log.Fatalf("Ошибка удаления таблиц: %v", err)
	}

	if err := models.Migrate(db); err != nil {
		log.Fatalf("Ошибка миграции базы данных: %v", err)
	}

	log.Println("База данных успешно сброшена!")
}

func runAdminSetup(db *gorm.DB) {
	// Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin789"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	// Создаем пользователя с email admin@example.com
	adminUser := models.User{
		Email:        "admin@example.com",
		PasswordHash: string(hashedPassword),
	}

	if err := db.Create(&adminUser).Error; err != nil {
		log.Fatalf("failed to create admin user: %v", err)
	}

	log.Println("Admin user created successfully!")
} */