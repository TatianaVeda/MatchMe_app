package controllers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"m/backend/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	AdminID       = "123e4567-e89b-12d3-a456-426614174000"
	AdminEmail    = "admin@first.av"
	AdminPassword = "qwaszx"
)

var fixturesDB *gorm.DB

func InitFixturesController(db *gorm.DB) {
	fixturesDB = db
	logrus.Info("Fixtures controller initialized")
}

func ResetFixtures(w http.ResponseWriter, r *http.Request) {
	// Сброс таблиц
	modelsToDrop := []interface{}{
		&models.User{}, &models.Profile{}, &models.Bio{}, &models.Preference{},
		&models.Recommendation{}, &models.Connection{}, &models.Chat{},
		&models.Message{}, &models.FakeUser{},
	}
	if err := fixturesDB.Migrator().DropTable(modelsToDrop...); err != nil {
		logrus.Errorf("ResetFixtures: ошибка удаления таблиц: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка удаления таблиц: %v", err), http.StatusInternalServerError)
		return
	}
	logrus.Info("ResetFixtures: таблицы успешно удалены")
	// Миграция БД заново
	if err := models.Migrate(fixturesDB); err != nil {
		logrus.Errorf("ResetFixtures: ошибка миграции БД: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка миграции базы данных: %v", err), http.StatusInternalServerError)
		return
	}
	logrus.Info("ResetFixtures: миграция БД выполнена успешно")

	adminUUID, _ := uuid.Parse(AdminID)
	var existing models.User
	if err := fixturesDB.First(&existing, "id = ?", adminUUID).Error; err == gorm.ErrRecordNotFound {
		hash, _ := models.HashPassword(AdminPassword)
		admin := models.User{
			ID:           adminUUID,
			Email:        AdminEmail,
			PasswordHash: hash,
		}
		if err := fixturesDB.Create(&admin).Error; err != nil {
			logrus.Errorf("ResetFixtures: не удалось создать админа: %v", err)
		} else {
			logrus.Infof("ResetFixtures: админ %s создан (ID=%s)", AdminEmail, AdminID)
		}
	} else {
		logrus.Info("ResetFixtures: админ уже существует, создание пропущено")
	}

	// Определение количества пользователей (по умолчанию 100, можно переопределить через ?num=)
	numUsers := 100
	if numParam := r.URL.Query().Get("num"); numParam != "" {
		if n, err := strconv.Atoi(numParam); err == nil && n > 0 {
			numUsers = n
		} else {
			logrus.Warnf("ResetFixtures: неверное значение параметра num (%s), используется значение по умолчанию 100", numParam)
		}
	}

	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Создание фиктивных пользователей с разнообразными профилями и биографиями
	for i := 1; i <= numUsers; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		password := "password123" // простой пароль для тестирования
		user, err := models.CreateUser(fixturesDB, email, password)
		if err != nil {
			logrus.Warnf("ResetFixtures: ошибка создания пользователя %s: %v", email, err)
			continue
		}

		// Создаем профиль с случайными данными
		profile := models.Profile{
			UserID:    user.ID,
			FirstName: randomFirstName(),
			LastName:  randomLastName(),
			About:     "Фиктивный пользователь для тестирования.",
			PhotoURL:  "/static/images/default.png", // Если фото не загружено, можно использовать placeholder
			Online:    false,
			Latitude:  randomLatitude(),
			Longitude: randomLongitude(),
		}
		// Создаем биографию с разнообразными данными (не менее 5 полей)
		bio := models.Bio{
			UserID:    user.ID,
			Interests: randomInterests(),
			Hobbies:   randomHobbies(),
			Music:     randomMusic(),
			Food:      randomFood(),
			Travel:    randomTravel(),
		}
		if err := fixturesDB.Save(&profile).Error; err != nil {
			logrus.Warnf("ResetFixtures: ошибка сохранения профиля для %s: %v", email, err)
		}
		if err := fixturesDB.Save(&bio).Error; err != nil {
			logrus.Warnf("ResetFixtures: ошибка сохранения биографии для %s: %v", email, err)
		}
		logrus.Debugf("ResetFixtures: пользователь %s создан", email)
	}
	logrus.Infof("ResetFixtures: загружено фиктивных пользователей: %d", numUsers)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("База данных сброшена и загружено %d фиктивных пользователей", numUsers),
	})
}

// --- Вспомогательные функции для генерации случайных данных ---

func randomFirstName() string {
	firstNames := []string{"Анна", "Иван", "Мария", "Алексей", "Ольга", "Дмитрий", "Елена", "Сергей", "Наталья", "Михаил"}
	return firstNames[rand.Intn(len(firstNames))]
}

func randomLastName() string {
	lastNames := []string{"Иванова", "Петров", "Сидорова", "Кузнецов", "Смирнова", "Морозов", "Новикова", "Фёдоров", "Соколова", "Михайлов"}
	return lastNames[rand.Intn(len(lastNames))]
}

func randomInterests() string {
	interests := []string{"кино", "спорт", "музыка", "технологии", "искусство", "путешествия", "литература", "фотография"}
	return interests[rand.Intn(len(interests))]
}

func randomHobbies() string {
	hobbies := []string{"чтение", "бег", "рисование", "игры", "готовка", "садоводство", "плавание", "путешествия"}
	return hobbies[rand.Intn(len(hobbies))]
}

func randomMusic() string {
	music := []string{"рок", "джаз", "классика", "поп", "хип-хоп", "электронная", "блюз"}
	return music[rand.Intn(len(music))]
}

func randomFood() string {
	foods := []string{"итальянская", "азиатская", "русская", "французская", "мексиканская", "японская"}
	return foods[rand.Intn(len(foods))]
}

func randomTravel() string {
	travel := []string{"пляжный отдых", "горный туризм", "культурные туры", "экспедиции", "городские экскурсии"}
	return travel[rand.Intn(len(travel))]
}

func randomLatitude() float64 {
	// Например, для России: от 41 до 82 градусов северной широты
	return 41 + rand.Float64()*(82-41)
}

// Suomi!!!
func randomLongitude() float64 {
	// Например, для России: от 19 до 169 градусов восточной долготы
	return 19 + rand.Float64()*(169-19)
}
