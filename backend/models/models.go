package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
	EarthLoc  []byte    `gorm:"type:cube;->" json:"-"`
}
type Bio struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`
	Interests  string    `gorm:"type:varchar(50)" json:"interests"`
	Hobbies    string    `gorm:"type:varchar(50)" json:"hobbies"`
	Music      string    `gorm:"type:varchar(50)" json:"music"`
	Food       string    `gorm:"type:varchar(50)" json:"food"`
	Travel     string    `gorm:"type:varchar(50)" json:"travel"`
	LookingFor string    `gorm:"type:text" json:"lookingFor"`
}
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
type Recommendation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	RecUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"recUserId"`
	Status    string    `gorm:"size:50;default:'pending'" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}
type Connection struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	ConnectionID uuid.UUID `gorm:"type:uuid;not null;index" json:"connectionId"`
	Status       string    `gorm:"size:50;not null" json:"status"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
}
type Chat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	User1ID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user1Id"`
	User2ID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user2Id"`
	Messages  []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE;" json:"messages"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    uint      `gorm:"not null;index" json:"chatId"`
	SenderID  uuid.UUID `gorm:"type:uuid;not null" json:"senderId"`
	Content   string    `gorm:"type:text" json:"content"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Read      bool      `gorm:"default:false" json:"read"`
	Sender    User      `json:"sender" gorm:"foreignKey:SenderID"`
}
type FakeUser struct {
	ID     uint      `gorm:"primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
}

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
	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; execErr != nil {
		logrus.Warnf("InitDB: не удалось создать extension uuid-ossp: %v", execErr)
	}
	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS cube`).Error; execErr != nil {
		logrus.Warnf("InitDB: не удалось создать extension cube: %v", execErr)
	}
	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS earthdistance`).Error; execErr != nil {
		logrus.Warnf("InitDB: не удалось создать extension earthdistance: %v", execErr)
	}
	if migrateErr := Migrate(db); migrateErr != nil {
		logrus.Errorf("InitDB: ошибка миграции: %v", migrateErr)
		return nil, migrateErr
	}
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
