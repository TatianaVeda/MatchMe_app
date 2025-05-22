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
	JWTExpiresIn        int
	JWTRefreshExpiresIn int
	MediaUploadDir      string
	Environment         string
	IsDev               bool
	IsProd              bool
	AllowedOrigins      []string
	SMTPServer          string
	SMTPPort            int
	SMTPUser            string
	SMTPPassword        string
	RedisURL            string
	RedisTimeout        int
	LogLevel            string
}

var AppConfig *Config

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
		SMTPServer:          getEnv("SMTP_SERVER", "smtp.example.com"),
		SMTPPort:            getEnvAsInt("SMTP_PORT", 587),
		SMTPUser:            getEnv("SMTP_USER", "user@example.com"),
		SMTPPassword:        getEnv("SMTP_PASSWORD", "password"),
		//RedisURL:            getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		RedisTimeout: getEnvAsInt("REDIS_TIMEOUT", 5),
		LogLevel:     getEnv("LOG_LEVEL", "debug"),
	}

	AppConfig.IsDev = AppConfig.Environment == "development"
	AppConfig.IsProd = AppConfig.Environment == "production"

	if err := AppConfig.Validate(); err != nil {
		logrus.Fatalf("Ошибка конфигурации: %v", err)
	}

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

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valStr := os.Getenv(name)
	if valStr == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		logrus.Warnf("Не удалось преобразовать %s в int: %v. Используется значение по умолчанию.", name, err)
		return defaultVal
	}

	return val
}
