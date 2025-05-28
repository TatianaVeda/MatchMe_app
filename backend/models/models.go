package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User represents an application user with authentication and profile data.
// It includes associations with Profile, Bio, and Preference models.
// All associations are configured with CASCADE delete for data consistency.
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

// Profile contains public user information and geolocation data.
// EarthLoc is a generated field using PostgreSQL's cube/earthdistance extensions
// for efficient geospatial queries and distance calculations.
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
	// EarthLoc is a generated column (PostgreSQL cube type) for fast geo-distance queries.
	// It is automatically computed from latitude/longitude using ll_to_earth().
	EarthLoc []byte `gorm:"type:cube;->" json:"-"`
}

// Bio contains user interests and search preferences for recommendations.
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

// Preference stores user search settings and field priorities for recommendations.
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

// Recommendation links a user to a recommended user and tracks status (pending/declined).
type Recommendation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	RecUserID uuid.UUID `gorm:"type:uuid;not null;index" json:"recUserId"`
	Status    string    `gorm:"size:50;default:'pending'" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// Connection represents a friendship or pending friend request between users.
type Connection struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
	ConnectionID uuid.UUID `gorm:"type:uuid;not null;index" json:"connectionId"`
	Status       string    `gorm:"size:50;not null" json:"status"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// Chat represents a chat between two users, with all messages as a slice.
type Chat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	User1ID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user1Id"`
	User2ID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user2Id"`
	Messages  []Message `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE;" json:"messages"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

// Message is a single message in a chat, with sender and timestamp.
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    uint      `gorm:"not null;index" json:"chatId"`
	SenderID  uuid.UUID `gorm:"type:uuid;not null" json:"senderId"`
	Content   string    `gorm:"type:text" json:"content"`
	Timestamp time.Time `gorm:"autoCreateTime" json:"timestamp"`
	Read      bool      `gorm:"default:false" json:"read"`
	Sender    User      `json:"sender" gorm:"foreignKey:SenderID"`
}

// FakeUser is used for marking test/dummy users in the database.
type FakeUser struct {
	ID     uint      `gorm:"primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"userId"`
}

// InitDB initializes the database connection using GORM (Go ORM library for SQL databases),
// runs migrations, and ensures required PostgreSQL extensions.
// Implements retry logic (up to 10 attempts) for robust startup in case of slow DB.
// Also ensures geo extensions (cube, earthdistance) and creates a generated column
// for geospatial queries (earth_loc), plus a GIST index for fast location search.
// All errors and progress are logged via logrus.
func InitDB(databaseURL string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	for i := 1; i <= 10; i++ {
		// Retry connection to handle cases when DB is not ready yet (e.g., in Docker Compose)
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
			logrus.Infof("InitDB: successfully connected to database (attempt %d)", i)
			break
		}
		logrus.Warnf("InitDB: attempt %d failed: %v", i, err)
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("InitDB: failed to connect after multiple attempts: %w", err)
	}
	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; execErr != nil {
		logrus.Warnf("InitDB: failed to create extension uuid-ossp: %v", execErr)
	}
	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS cube`).Error; execErr != nil {
		logrus.Warnf("InitDB: failed to create extension cube: %v", execErr)
	}
	if execErr := db.Exec(`CREATE EXTENSION IF NOT EXISTS earthdistance`).Error; execErr != nil {
		logrus.Warnf("InitDB: failed to create extension earthdistance: %v", execErr)
	}
	if migrateErr := Migrate(db); migrateErr != nil {
		logrus.Errorf("InitDB: migration error: %v", migrateErr)
		return nil, migrateErr
	}
	// Add generated column earth_loc for geospatial queries (ll_to_earth)
	if err := db.Exec(`
		ALTER TABLE profiles
		ADD COLUMN IF NOT EXISTS earth_loc cube
		GENERATED ALWAYS AS (ll_to_earth(latitude, longitude)) STORED
	`).Error; err != nil {
		logrus.Warnf("InitDB: failed to add earth_loc column: %v", err)
	}
	// Create GIST index for fast geo-search by earth_loc
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_profiles_earth_loc
		ON profiles USING GIST (earth_loc)
	`).Error; err != nil {
		logrus.Warnf("InitDB: failed to create index idx_profiles_earth_loc: %v", err)
	}
	logrus.Info("InitDB: database initialized successfully")
	return db, nil
}

// Migrate runs GORM automigrations for all main models and logs the result.
// Ensures all tables and relations are up to date with Go structs.
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
		logrus.Errorf("Migrate: migration error: %v", err)
	} else {
		logrus.Info("Migrate: migration completed successfully")
	}
	return err
}
