package controllers

import (
	"encoding/json"
	"errors"
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
	AdminPassword = "qwaszx12"
)

var fixturesDB *gorm.DB

func InitFixturesController(db *gorm.DB) {
	fixturesDB = db
	logrus.Info("Fixtures controller initialized")
}

// ResetFixtures удаляет все таблицы, мигрирует схему заново и создаёт администратора.
func ResetFixtures(w http.ResponseWriter, r *http.Request) {
	// 1) Сброс всех таблиц
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

	// 2) Миграция схемы
	if err := models.Migrate(fixturesDB); err != nil {
		logrus.Errorf("ResetFixtures: ошибка миграции БД: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка миграции базы данных: %v", err), http.StatusInternalServerError)
		return
	}
	logrus.Info("ResetFixtures: миграция БД выполнена успешно")

	// 3) Создание (или проверка существования) администратора
	adminUUID, _ := uuid.Parse(AdminID)
	var existing models.User
	err := fixturesDB.First(&existing, "id = ?", adminUUID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "База данных сброшена, администратор сохранён",
	})
}

// GenerateFixtures создаёт указанное число фейковых пользователей, администратора не затрагивая.
func GenerateFixtures(w http.ResponseWriter, r *http.Request) {
	// По умолчанию 100 пользователей, но можно переопределить через ?num=
	numUsers := 100
	if param := r.URL.Query().Get("num"); param != "" {
		if n, err := strconv.Atoi(param); err == nil && n > 0 {
			numUsers = n
		} else {
			logrus.Warnf("GenerateFixtures: неверное значение параметра num (%s), используется %d", param, numUsers)
		}
	}

	rand.Seed(time.Now().UnixNano())

	for i := 1; i <= numUsers; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		password := "password123"
		user, err := models.CreateUser(fixturesDB, email, password)
		if err != nil {
			logrus.Warnf("GenerateFixtures: ошибка создания пользователя %s: %v", email, err)
			continue
		}

		profile := models.Profile{
			UserID:    user.ID,
			FirstName: randomFirstName(),
			LastName:  randomLastName(),
			About:     "Фиктивный пользователь для тестирования.",
			PhotoURL:  "/static/images/default.png",
			Online:    false,
			Latitude:  randomLatitude(),
			Longitude: randomLongitude(),
		}
		bio := models.Bio{
			UserID:    user.ID,
			Interests: randomInterests(),
			Hobbies:   randomHobbies(),
			Music:     randomMusic(),
			Food:      randomFood(),
			Travel:    randomTravel(),
		}

		if err := fixturesDB.Save(&profile).Error; err != nil {
			logrus.Warnf("GenerateFixtures: ошибка сохранения профиля %s: %v", email, err)
		}
		if err := fixturesDB.Save(&bio).Error; err != nil {
			logrus.Warnf("GenerateFixtures: ошибка сохранения биографии %s: %v", email, err)
		}

		logrus.Debugf("GenerateFixtures: создан пользователь %s", email)
	}

	logrus.Infof("GenerateFixtures: создано %d фейковых пользователей", numUsers)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Сгенерировано %d фейковых пользователей", numUsers),
	})
}

// --- Вспомогательные функции для случайных данных ---

func randomFirstName() string {
	arr := []string{"Анна", "Иван", "Мария", "Алексей", "Ольга", "Дмитрий", "Елена", "Сергей", "Наталья", "Михаил"}
	return arr[rand.Intn(len(arr))]
}

func randomLastName() string {
	arr := []string{"Иванова", "Петров", "Сидорова", "Кузнецов", "Смирнова", "Морозов", "Новикова", "Фёдоров", "Соколова", "Михайлов"}
	return arr[rand.Intn(len(arr))]
}

func randomInterests() string {
	arr := []string{"кино", "спорт", "музыка", "технологии", "искусство", "путешествия", "литература", "фотография"}
	return arr[rand.Intn(len(arr))]
}

func randomHobbies() string {
	arr := []string{"чтение", "бег", "рисование", "игры", "готовка", "садоводство", "плавание", "путешествия"}
	return arr[rand.Intn(len(arr))]
}

func randomMusic() string {
	arr := []string{"рок", "джаз", "классика", "поп", "хип-хоп", "электронная", "блюз"}
	return arr[rand.Intn(len(arr))]
}

func randomFood() string {
	arr := []string{"итальянская", "азиатская", "русская", "французская", "мексиканская", "японская"}
	return arr[rand.Intn(len(arr))]
}

func randomTravel() string {
	arr := []string{"пляжный отдых", "горный туризм", "культурные туры", "экспедиции", "городские экскурсии"}
	return arr[rand.Intn(len(arr))]
}

func randomLatitude() float64 {
	return 41 + rand.Float64()*(82-41)
}

func randomLongitude() float64 {
	return 19 + rand.Float64()*(169-19)
}
