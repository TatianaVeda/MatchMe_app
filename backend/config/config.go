package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type Config struct {
	ServerPort          string
	WebSocketPort       string
	DatabaseURL         string
	JWTSecret           string
	JWTExpiresIn        int // время в минутах
	JWTRefreshExpiresIn int // время в минутах
	MediaUploadDir      string
	Environment         string
	IsDev               bool
	IsProd              bool
	AllowedOrigins      []string // для CORS

	// Новые параметры:
	SMTPServer   string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	RedisURL     string
	RedisTimeout int // секунды

	LogLevel string // debug, info, warn, error
}

var AppConfig *Config

// LoadConfig загружает конфигурацию из переменных окружения.
func LoadConfig() {
	AppConfig = &Config{
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		WebSocketPort:       getEnv("WEBSOCKET_PORT", "8081"),
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://user:pass@localhost:5433/sopostavmenya?sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "supersecretjwtkey"),
		JWTExpiresIn:        getEnvAsInt("JWT_EXPIRES_IN", 60),
		JWTRefreshExpiresIn: getEnvAsInt("JWT_REFRESH_EXPIRES_IN", 10080),
		MediaUploadDir:      getEnv("MEDIA_UPLOAD_DIR", "./static/images"),
		Environment:         strings.ToLower(getEnv("ENVIRONMENT", "development")),
		AllowedOrigins:      strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080"), ","),

		// Новые параметры:
		SMTPServer:   getEnv("SMTP_SERVER", "smtp.example.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", "user@example.com"),
		SMTPPassword: getEnv("SMTP_PASSWORD", "password"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisTimeout: getEnvAsInt("REDIS_TIMEOUT", 5),
		LogLevel:     getEnv("LOG_LEVEL", "debug"),
	}

	// Автоопределяем режим
	AppConfig.IsDev = AppConfig.Environment == "development"
	AppConfig.IsProd = AppConfig.Environment == "production"

	// Валидация конфигурации
	if err := AppConfig.Validate(); err != nil {
		logrus.Fatalf("Ошибка конфигурации: %v", err)
	}

	// Настройка логирования
	level, err := logrus.ParseLevel(AppConfig.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	if AppConfig.IsDev {
		logrus.SetReportCaller(true)
		logrus.Infof("Режим разработки: детальное логирование включено")
	}

	logrus.Info("✅ Конфигурация загружена")
}

// Validate выполняет базовую валидацию значений конфигурации.
func (c *Config) Validate() error {
	if c.ServerPort == "" {
		return errors.New("SERVER_PORT не может быть пустым")
	}
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL не может быть пустым")
	}
	if c.JWTSecret == "" {
		return errors.New("JWT_SECRET не может быть пустым")
	}
	if c.JWTExpiresIn <= 0 {
		return errors.New("JWT_EXPIRES_IN должен быть больше нуля")
	}
	if len(c.AllowedOrigins) == 0 {
		return errors.New("ALLOWED_ORIGINS должны быть указаны")
	}
	if c.SMTPServer == "" {
		return errors.New("SMTP_SERVER не может быть пустым")
	}
	if c.SMTPPort <= 0 {
		return errors.New("SMTP_PORT должен быть больше нуля")
	}
	if c.LogLevel == "" {
		return errors.New("LOG_LEVEL не может быть пустым")
	}
	return nil
}

// getEnv возвращает значение переменной окружения или значение по умолчанию.
func getEnv(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

// getEnvAsInt возвращает значение переменной окружения как int или значение по умолчанию.
func getEnvAsInt(name string, defaultVal int) int {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		logrus.Warnf("❌ Не удалось преобразовать %s в int: %v. Используется значение по умолчанию.", name, err)
		return defaultVal
	}
	return val
}
