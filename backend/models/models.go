package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ПРОВЕРИТЬ
// func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
// 	if u.ID == uuid.Nil {
// 		u.ID = uuid.New()
// 	}
// 	return
// }

// User представляет таблицу пользователей.
type User struct {
	// Используем UUID в качестве первичного ключа.
	//ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"unique;not null" json:"-"` // email не будет сериализован в JSON
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Связанные записи (один к одному)
	Profile    Profile    `gorm:"constraint:OnDelete:CASCADE;" json:"profile"`
	Bio        Bio        `gorm:"constraint:OnDelete:CASCADE;" json:"bio"`
	Preference Preference `gorm:"constraint:OnDelete:CASCADE;" json:"preference"`
}

// Profile представляет информацию «Обо мне» и связанные данные профиля.
type Profile struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Внешний ключ для связи с пользователем.
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	// Данные профиля.
	FirstName string  `gorm:"size:255" json:"first_name"`
	LastName  string  `gorm:"size:255" json:"last_name"`
	About     string  `gorm:"type:text" json:"about"`
	PhotoURL  string  `gorm:"size:512" json:"photo_url"`   // Ссылка или путь к изображению.
	Online    bool    `gorm:"default:false" json:"online"` // Индикатор онлайн/офлайн.
	Latitude  float64 `json:"latitude"`                    // Координаты для фильтрации по местоположению.
	Longitude float64 `json:"longitude"`                   //`json:"longitude"`
}

// Bio хранит дополнительные биографические данные для рекомендаций.
// Здесь задаются не менее пяти полей: интересы, хобби, музыкальные и кулинарные предпочтения, путешествия.
type Bio struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	Interests string `gorm:"type:text" json:"interests"`
	Hobbies   string `gorm:"type:text" json:"hobbies"`
	Music     string `gorm:"type:text" json:"music"`
	Food      string `gorm:"type:text" json:"food"`
	Travel    string `gorm:"type:text" json:"travel"`
}

// Preference хранит настройки поиска пользователя, например максимальный радиус рекомендаций.
type Preference struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	// Максимальный радиус для рекомендаций (например, в километрах).
	MaxRadius float64 `gorm:"default:0" json:"max_radius"`
}

// Recommendation хранит информацию о рекомендациях,
// чтобы избежать повторного показа уже отклонённых или просмотренных профилей.
type Recommendation struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Пользователь, которому предлагается рекомендация.
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	// Рекомендуемый пользователь.
	RecUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"rec_user_id"`

	// Статус рекомендации: например, "pending", "rejected" и т.д.
	Status string `gorm:"size:50;default:'pending'" json:"status"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Connection описывает связь между пользователями.
type Connection struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Инициатор запроса.
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	// Пользователь, с которым устанавливается связь.
	ConnectionID uuid.UUID `gorm:"type:uuid;not null;index" json:"connection_id"`

	// Статус запроса: "requested", "accepted", "rejected".
	Status string `gorm:"size:50;not null" json:"status"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Chat представляет чат между двумя пользователями.
// Гарантируется, что между двумя пользователями существует только одна запись чата.
type Chat struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Идентификаторы пользователей, участвующих в чате.
	User1ID uuid.UUID `gorm:"type:uuid;not null;index" json:"user1_id"`
	User2ID uuid.UUID `gorm:"type:uuid;not null;index" json:"user2_id"`

	// Сообщения, связанные с этим чатом.
	Messages []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE;" json:"messages"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Message представляет отдельное сообщение в чате.
type Message struct {
	ID uint `gorm:"primaryKey" json:"id"`

	ChatID uint `gorm:"not null;index" json:"chat_id"`
	// Отправитель сообщения.
	SenderID uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	Content  string    `gorm:"type:text" json:"content"`

	// Время отправки сообщения.
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
	// Флаг прочтения сообщения.
	Read bool `gorm:"default:false" json:"read"`
}

// FakeUser используется для фиктивных пользователей (например, для загрузки тестовых данных).
// Можно использовать отдельную таблицу или добавить флаг в таблицу users.
type FakeUser struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// Связанный пользователь.
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	// Дополнительные поля, если необходимо.
}

// InitDB инициализирует подключение к базе данных PostgreSQL с использованием GORM.
func InitDB(databaseURL string) (*gorm.DB, error) {
	// Если используется расширение uuid-ossp, можно создать его при инициализации:
	// db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		logrus.Errorf("InitDB: ошибка подключения к базе данных: %v", err)
		return nil, err
	}

	// Только для PostgreSQL
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	// Выполняем миграцию всех моделей.
	if err := Migrate(db); err != nil {
		logrus.Errorf("InitDB: ошибка миграции базы данных: %v", err)
		return nil, err
	}
	logrus.Info("InitDB: база данных успешно инициализирована")
	return db, nil
}

// Migrate выполняет автоматическую миграцию для всех моделей.
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&User{},
		&Profile{},
		&Bio{},
		&Preference{},
		&Recommendation{},
		&Connection{},
		&Chat{},
		&Message{},
		&FakeUser{},
	)
	if err != nil {
		logrus.Errorf("Migrate: ошибка миграции: %v", err)
	} else {
		logrus.Info("Migrate: миграция выполнена успешно")
	}
	return err
}
