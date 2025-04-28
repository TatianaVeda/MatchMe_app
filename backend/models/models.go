// package models

// import (
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/sirupsen/logrus"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// // ПРОВЕРИТЬ
// // func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
// // 	if u.ID == uuid.Nil {
// // 		u.ID = uuid.New()
// // 	}
// // 	return
// // }

// // User представляет таблицу пользователей.
// type User struct {
// 	// Используем UUID в качестве первичного ключа.
// 	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
// 	//ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
// 	Email        string    `gorm:"unique;not null" json:"-"` // email не будет сериализован в JSON
// 	PasswordHash string    `gorm:"not null" json:"-"`
// 	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
// 	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

// 	// Связанные записи (один к одному)
// 	Profile    Profile    `gorm:"constraint:OnDelete:CASCADE;" json:"profile"`
// 	Bio        Bio        `gorm:"constraint:OnDelete:CASCADE;" json:"bio"`
// 	Preference Preference `gorm:"constraint:OnDelete:CASCADE;" json:"preference"`
// }

// // Profile представляет информацию «Обо мне» и связанные данные профиля.
// type Profile struct {
// 	ID uint `gorm:"primaryKey" json:"id"`

// 	// Внешний ключ для связи с пользователем.
// 	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`

// 	// Данные профиля.
// 	FirstName string  `gorm:"size:255" json:"firstName"`
// 	LastName  string  `gorm:"size:255" json:"lastName"`
// 	About     string  `gorm:"type:text" json:"about"`
// 	PhotoURL  string  `gorm:"size:512" json:"photoUrl"`    // Ссылка или путь к изображению.
// 	Online    bool    `gorm:"default:false" json:"online"` // Индикатор онлайн/офлайн.
// 	Latitude  float64 `json:"latitude"`                    // Координаты для фильтрации по местоположению.
// 	Longitude float64 `json:"longitude"`                   //`json:"longitude"`
// }

// // Bio хранит дополнительные биографические данные для рекомендаций.
// // Здесь задаются не менее пяти полей: интересы, хобби, музыкальные и кулинарные предпочтения, путешествия.
// type Bio struct {
// 	ID uint `gorm:"primaryKey" json:"id"`

// 	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`

// 	Interests string `gorm:"type:text" json:"interests"`
// 	Hobbies   string `gorm:"type:text" json:"hobbies"`
// 	Music     string `gorm:"type:text" json:"music"`
// 	Food      string `gorm:"type:text" json:"food"`
// 	Travel    string `gorm:"type:text" json:"travel"`
// }

// // Preference хранит настройки поиска пользователя, например максимальный радиус рекомендаций.
// type Preference struct {
// 	ID uint `gorm:"primaryKey" json:"id"`

// 	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`

// 	// Максимальный радиус для рекомендаций (например, в километрах).
// 	MaxRadius float64 `gorm:"default:0" json:"maxRadius"`
// }

// // Recommendation хранит информацию о рекомендациях,
// // чтобы избежать повторного показа уже отклонённых или просмотренных профилей.
// type Recommendation struct {
// 	ID uint `gorm:"primaryKey" json:"id"`

// 	// Пользователь, которому предлагается рекомендация.
// 	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
// 	// Рекомендуемый пользователь.
// 	RecUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"recUserId"`

// 	// Статус рекомендации: например, "pending", "rejected" и т.д.
// 	Status string `gorm:"size:50;default:'pending'" json:"status"`

// 	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
// }

// // Connection описывает связь между пользователями.
// type Connection struct {
// 	ID uint `gorm:"primaryKey" json:"id"`

// 	// Инициатор запроса.
// 	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
// 	// Пользователь, с которым устанавливается связь.
// 	ConnectionID uuid.UUID `gorm:"type:uuid;not null;index" json:"connectionId"`

// 	// Статус запроса: "requested", "accepted", "rejected".
// 	Status string `gorm:"size:50;not null" json:"status"`

// 	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
// }

// // Chat представляет чат между двумя пользователями.
// // Гарантируется, что между двумя пользователями существует только одна запись чата.
// type Chat struct {
// 	ID uint `gorm:"primaryKey" json:"id"`

// 	// Идентификаторы пользователей, участвующих в чате.
// 	User1ID uuid.UUID `gorm:"type:uuid;not null;index" json:"user1Id"`
// 	User2ID uuid.UUID `gorm:"type:uuid;not null;index" json:"user2Id"`

// 	// Сообщения, связанные с этим чатом.
// 	Messages []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE;" json:"messages"`

// 	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
// }

// // Message представляет отдельное сообщение в чате.
// type Message struct {
// 	ID uint `gorm:"primaryKey" json:"id"`

// 	ChatID uint `gorm:"not null;index" json:"chatId"`
// 	// Отправитель сообщения.
// 	SenderID uuid.UUID `gorm:"type:uuid;not null" json:"senderId"`
// 	Content  string    `gorm:"type:text" json:"content"`

// 	// Время отправки сообщения.
// 	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
// 	// Флаг прочтения сообщения.
// 	Read bool `gorm:"default:false" json:"read"`
// }

// // FakeUser используется для фиктивных пользователей (например, для загрузки тестовых данных).
// // Можно использовать отдельную таблицу или добавить флаг в таблицу users.
// type FakeUser struct {
// 	ID uint `gorm:"primaryKey" json:"id"`
// 	// Связанный пользователь.
// 	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
// 	// Дополнительные поля, если необходимо.
// }

// // InitDB инициализирует подключение к базе данных PostgreSQL с использованием GORM.
// func InitDB(databaseURL string) (*gorm.DB, error) {
// 	// Если используется расширение uuid-ossp, можно создать его при инициализации:
// 	// db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
// 	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
// 	if err != nil {
// 		logrus.Errorf("InitDB: ошибка подключения к базе данных: %v", err)
// 		return nil, err
// 	}

// 	// Только для PostgreSQL
// 	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

// 	// Выполняем миграцию всех моделей.
// 	if err := Migrate(db); err != nil {
// 		logrus.Errorf("InitDB: ошибка миграции базы данных: %v", err)
// 		return nil, err
// 	}
// 	logrus.Info("InitDB: база данных успешно инициализирована")
// 	return db, nil
// }

// // Migrate выполняет автоматическую миграцию для всех моделей.
// func Migrate(db *gorm.DB) error {
// 	err := db.AutoMigrate(
// 		&User{},
// 		&Profile{},
// 		&Bio{},
// 		&Preference{},
// 		&Recommendation{},
// 		&Connection{},
// 		&Chat{},
// 		&Message{},
// 		&FakeUser{},
// 	)
// 	if err != nil {
// 		logrus.Errorf("Migrate: ошибка миграции: %v", err)
// 	} else {
// 		logrus.Info("Migrate: миграция выполнена успешно")
// 	}
// 	return err
// }

// package models

// import (
// 	"fmt"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/sirupsen/logrus"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// // User представляет таблицу пользователей.
// type User struct {
// 	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
// 	Email        string    `gorm:"unique;not null" json:"-"`
// 	PasswordHash string    `gorm:"not null" json:"-"`
// 	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
// 	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

// 	Profile    Profile    `gorm:"constraint:OnDelete:CASCADE;" json:"profile"`
// 	Bio        Bio        `gorm:"constraint:OnDelete:CASCADE;" json:"bio"`
// 	Preference Preference `gorm:"constraint:OnDelete:CASCADE;" json:"preference"`
// }

// // Profile представляет информацию "Обо мне" и связанные данные профиля.
// type Profile struct {
// 	ID        uint      `gorm:"primaryKey" json:"id"`
// 	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
// 	FirstName string    `gorm:"size:255" json:"firstName"`
// 	LastName  string    `gorm:"size:255" json:"lastName"`
// 	About     string    `gorm:"type:text" json:"about"`
// 	PhotoURL  string    `gorm:"size:512" json:"photoUrl"`
// 	Online    bool      `gorm:"default:false" json:"online"`
// 	City      string    `gorm:"size:100" json:"city"` // новое поле «город»
// 	Latitude  float64   `json:"latitude"`
// 	Longitude float64   `json:"longitude"`
// 	EarthLoc  []byte    `gorm:"type:cube;->" json:"-"` // Read-only, заполняется в БД
// }

// // Bio хранит дополнительные биографические данные.
// type Bio struct {
// 	ID         uint      `gorm:"primaryKey" json:"id"`
// 	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
// 	Interests  string    `gorm:"type:varchar(50)" json:"interests"` // one of [Psychology, Ecology, Philosophy, Space and Astronomy, Fashion and Design]
// 	Hobbies    string    `gorm:"type:varchar(50)" json:"hobbies"`   // one of [Sports, Nature, Art, Handicrafts, Making, IT]
// 	Music      string    `gorm:"type:varchar(50)" json:"music"`     // one of [Rock, Classical, Jazz, Pop, Electronic]
// 	Food       string    `gorm:"type:varchar(50)" json:"food"`      // one of [Italian, Japanese, Mexican, Vegan, French]
// 	Travel     string    `gorm:"type:varchar(50)" json:"travel"`    // one of [Beach vacations, Mountain hiking, Cultural city tours, Cruises, Eco-tourism]
// 	LookingFor string    `gorm:"type:text" json:"lookingFor"`
// }

// // Preference хранит настройки поиска пользователя.
// type Preference struct {
// 	ID        uint      `gorm:"primaryKey" json:"id"`
// 	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
// 	MaxRadius float64   `gorm:"default:0" json:"maxRadius"`
// 	// New priority flags: если true — совпадение в этой категории считается за 40%, иначе за 20%
// 	PriorityInterests bool `gorm:"default:false" json:"priorityInterests"`
// 	PriorityHobbies   bool `gorm:"default:false" json:"priorityHobbies"`
// 	PriorityMusic     bool `gorm:"default:false" json:"priorityMusic"`
// 	PriorityFood      bool `gorm:"default:false" json:"priorityFood"`
// 	PriorityTravel    bool `gorm:"default:false" json:"priorityTravel"`
// }

// // Recommendation хранит информацию о рекомендациях.
// type Recommendation struct {
// 	ID        uint      `gorm:"primaryKey" json:"id"`
// 	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
// 	RecUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"recUserId"`
// 	Status    string    `gorm:"size:50;default:'pending'" json:"status"`
// 	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
// }

// // Connection описывает связь между пользователями.
// type Connection struct {
// 	ID           uint      `gorm:"primaryKey" json:"id"`
// 	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
// 	ConnectionID uuid.UUID `gorm:"type:uuid;not null;index" json:"connectionId"`
// 	Status       string    `gorm:"size:50;not null" json:"status"`
// 	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
// }

// // Chat представляет чат между двумя пользователями.
// type Chat struct {
// 	ID        uint      `gorm:"primaryKey" json:"id"`
// 	User1ID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user1Id"`
// 	User2ID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user2Id"`
// 	Messages  []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE;" json:"messages"`
// 	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
// }

// // Message представляет отдельное сообщение в чате.
// type Message struct {
// 	ID        uint      `gorm:"primaryKey" json:"id"`
// 	ChatID    uint      `gorm:"not null;index" json:"chatId"`
// 	SenderID  uuid.UUID `gorm:"type:uuid;not null" json:"senderId"`
// 	Content   string    `gorm:"type:text" json:"content"`
// 	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
// 	Read      bool      `gorm:"default:false" json:"read"`
// }

// // FakeUser используется для фиктивных пользователей.
// type FakeUser struct {
// 	ID     uint      `gorm:"primaryKey" json:"id"`
// 	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
// }

// // InitDB инициализирует подключение к базе данных PostgreSQL с retry-механикой.
// func InitDB(databaseURL string) (*gorm.DB, error) {
// 	var db *gorm.DB
// 	var err error
// 	// Попытаться подключиться до 10 раз, с паузой в 1 секунду
// 	for i := 1; i <= 10; i++ {
// 		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
// 		if err == nil {
// 			// Выполняем Ping для проверки соединения
// 			sqlDB, pingErr := db.DB()
// 			if pingErr != nil {
// 				err = pingErr
// 			} else {
// 				err = sqlDB.Ping()
// 			}
// 		}
// 		if err == nil {
// 			logrus.Infof("InitDB: успешно подключились к базе данных (попытка %d)", i)
// 			break
// 		}
// 		logrus.Warnf("InitDB: попытка %d не удалась: %v", i, err)
// 		time.Sleep(time.Second)
// 	}
// 	if err != nil {
// 		return nil, fmt.Errorf("InitDB: не удалось подключиться после нескольких попыток: %w", err)
// 	}

// 	// Создаем расширение uuid-ossp, если нужно
// 	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; execErr != nil {
// 		logrus.Warnf("InitDB: не удалось создать extension uuid-ossp: %v", execErr)
// 	}

// 	// Выполняем миграцию всех моделей
// 	if migrateErr := Migrate(db); migrateErr != nil {
// 		logrus.Errorf("InitDB: ошибка миграции: %v", migrateErr)
// 		return nil, migrateErr
// 	}

// 	logrus.Info("InitDB: база данных успешно инициализирована")
// 	return db, nil
// }

// // Migrate выполняет автоматическую миграцию для всех моделей.
// func Migrate(db *gorm.DB) error {
// 	err := db.AutoMigrate(
// 		&User{},
// 		&Profile{},
// 		&Bio{},
// 		&Preference{},
// 		&Recommendation{},
// 		&Connection{},
// 		&Chat{},
// 		&Message{},
// 		&FakeUser{},
// 	)
// 	if err != nil {
// 		logrus.Errorf("Migrate: ошибка миграции: %v", err)
// 	} else {
// 		logrus.Info("Migrate: миграция выполнена успешно")
// 	}
// 	return err
// }

package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User представляет таблицу пользователей.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email        string    `gorm:"unique;not null" json:"-"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	Profile    Profile    `gorm:"constraint:OnDelete:CASCADE;" json:"profile"`
	Bio        Bio        `gorm:"constraint:OnDelete:CASCADE;" json:"bio"`
	Preference Preference `gorm:"constraint:OnDelete:CASCADE;" json:"preference"`
}

// Profile представляет информацию "Обо мне" и связанные данные профиля.
type Profile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	FirstName string    `gorm:"size:255" json:"firstName"`
	LastName  string    `gorm:"size:255" json:"lastName"`
	About     string    `gorm:"type:text" json:"about"`
	PhotoURL  string    `gorm:"size:512" json:"photoUrl"`
	Online    bool      `gorm:"default:false" json:"online"`
	City      string    `gorm:"size:100" json:"city"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	EarthLoc  []byte    `gorm:"type:cube;->" json:"-"` // Новое поле только для чтения
}

// Bio хранит дополнительные биографические данные.
type Bio struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	Interests  string    `gorm:"type:varchar(50)" json:"interests"`
	Hobbies    string    `gorm:"type:varchar(50)" json:"hobbies"`
	Music      string    `gorm:"type:varchar(50)" json:"music"`
	Food       string    `gorm:"type:varchar(50)" json:"food"`
	Travel     string    `gorm:"type:varchar(50)" json:"travel"`
	LookingFor string    `gorm:"type:text" json:"lookingFor"`
}

// Preference хранит настройки поиска пользователя.
type Preference struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UserID            uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	MaxRadius         float64   `gorm:"default:0" json:"maxRadius"`
	PriorityInterests bool      `gorm:"default:false" json:"priorityInterests"`
	PriorityHobbies   bool      `gorm:"default:false" json:"priorityHobbies"`
	PriorityMusic     bool      `gorm:"default:false" json:"priorityMusic"`
	PriorityFood      bool      `gorm:"default:false" json:"priorityFood"`
	PriorityTravel    bool      `gorm:"default:false" json:"priorityTravel"`
}

// Recommendation хранит информацию о рекомендациях.
type Recommendation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	RecUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"recUserId"`
	Status    string    `gorm:"size:50;default:'pending'" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// Connection описывает связь между пользователями.
type Connection struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	ConnectionID uuid.UUID `gorm:"type:uuid;not null;index" json:"connectionId"`
	Status       string    `gorm:"size:50;not null" json:"status"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// Chat представляет чат между двумя пользователями.
type Chat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	User1ID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user1Id"`
	User2ID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user2Id"`
	Messages  []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE;" json:"messages"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// Message представляет отдельное сообщение в чате.
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    uint      `gorm:"not null;index" json:"chatId"`
	SenderID  uuid.UUID `gorm:"type:uuid;not null" json:"senderId"`
	Content   string    `gorm:"type:text" json:"content"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Read      bool      `gorm:"default:false" json:"read"`
}

// FakeUser используется для фиктивных пользователей.
type FakeUser struct {
	ID     uint      `gorm:"primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
}

// InitDB инициализирует подключение к базе данных PostgreSQL с retry-механикой.
func InitDB(databaseURL string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 1; i <= 10; i++ {
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err == nil {
			sqlDB, pingErr := db.DB()
			if pingErr != nil {
				err = pingErr
			} else {
				err = sqlDB.Ping()
			}
		}
		if err == nil {
			logrus.Infof("InitDB: успешно подключились к базе данных (попытка %d)", i)
			break
		}
		logrus.Warnf("InitDB: попытка %d не удалась: %v", i, err)
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("InitDB: не удалось подключиться после нескольких попыток: %w", err)
	}

	// Создаем расширение uuid-ossp и earthdistance/cube
	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; execErr != nil {
		logrus.Warnf("InitDB: не удалось создать extension uuid-ossp: %v", execErr)
	}
	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS cube`).Error; execErr != nil {
		logrus.Warnf("InitDB: не удалось создать extension cube: %v", execErr)
	}
	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS earthdistance`).Error; execErr != nil {
		logrus.Warnf("InitDB: не удалось создать extension earthdistance: %v", execErr)
	}

	// Выполняем миграцию всех моделей
	if migrateErr := Migrate(db); migrateErr != nil {
		logrus.Errorf("InitDB: ошибка миграции: %v", migrateErr)
		return nil, migrateErr
	}

	// Дополнительно: создаём поле earth_loc и индекс, если их нет
	if err := db.Exec(`
		ALTER TABLE profiles
		ADD COLUMN IF NOT EXISTS earth_loc cube
		GENERATED ALWAYS AS (ll_to_earth(latitude, longitude)) STORED
	`).Error; err != nil {
		logrus.Warnf("InitDB: не удалось добавить поле earth_loc: %v", err)
	}
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_profiles_earth_loc
		ON profiles USING GIST (earth_loc)
	`).Error; err != nil {
		logrus.Warnf("InitDB: не удалось создать индекс idx_profiles_earth_loc: %v", err)
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
