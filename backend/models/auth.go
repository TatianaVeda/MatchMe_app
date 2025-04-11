package models

import (
	"errors"
	"time"

	"backend/utils"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/sirupsen/logrus"
)

// Ошибки для обработки аутентификации.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// HashPassword возвращает bcrypt-хэш для переданного пароля.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Errorf("HashPassword: ошибка генерации хэша: %v", err)
		return "", err
	}
	logrus.Debug("HashPassword: пароль успешно захэширован")
	return string(bytes), nil
}

// CheckPasswordHash сравнивает пароль с его bcrypt-хэшем.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		logrus.Debugf("CheckPasswordHash: не совпадает: %v", err)
		return false
	}
	return true
}

// CreateUser регистрирует нового пользователя, выполняя валидацию email и пароля,
// хэширование пароля и создание записи в БД.
func CreateUser(db *gorm.DB, email, password string) (*User, error) {
	if err := utils.ValidateEmail(email); err != nil {
		logrus.Warnf("CreateUser: неверный формат email %s: %v", email, err)
		return nil, err
	}
	if err := utils.ValidatePassword(password); err != nil {
		logrus.Warnf("CreateUser: пароль не соответствует требованиям: %v", err)
		return nil, err
	}

	var count int64
	if err := db.Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		logrus.Errorf("CreateUser: ошибка проверки существования email %s: %v", email, err)
		return nil, err
	}
	if count > 0 {
		logrus.Warnf("CreateUser: пользователь с данным email уже зарегистрирован")
		return nil, errors.New("email already registered")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
	}
	user.Profile = Profile{}
	user.Bio = Bio{}
	user.Preference = Preference{}

	if err := db.Create(user).Error; err != nil {
		logrus.Errorf("CreateUser: ошибка создания пользователя %s: %v", email, err)
		return nil, err
	}

	logrus.Infof("CreateUser: пользователь успешно создан")
	return user, nil
}

// AuthenticateUser выполняет аутентификацию, сравнивая переданный пароль с сохранённым хэшем.
// При успешном сравнении возвращает пользователя.
func AuthenticateUser(db *gorm.DB, email, password string) (*User, error) {
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Warnf("AuthenticateUser: пользователь с email %s не найден", email)
			return nil, ErrUserNotFound
		}
		logrus.Errorf("AuthenticateUser: ошибка поиска пользователя %s: %v", email, err)
		return nil, err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		logrus.Warnf("AuthenticateUser: неверный пароль для пользователя %s", email)
		return nil, ErrInvalidCredentials
	}

	logrus.Infof("AuthenticateUser: пользователь %s успешно аутентифицирован", email)
	return &user, nil
}

// JWTClaims определяет полезную нагрузку для JWT-токена.
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT генерирует JWT-токен для аутентифицированного пользователя.
func GenerateJWT(userID uuid.UUID, secret string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		logrus.Errorf("GenerateJWT: ошибка создания токена для пользователя %s: %v", userID, err)
		return "", err
	}

	logrus.Infof("GenerateJWT: токен успешно создан для пользователя %s", userID)
	return tokenString, nil
}

// ParseJWT проверяет JWT-токен и возвращает его claims.
func ParseJWT(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		logrus.Errorf("ParseJWT: ошибка при парсинге токена: %v", err)
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		logrus.Warn("ParseJWT: неверные claims или токен не валиден")
		return nil, errors.New("invalid token")
	}

	logrus.Debug("ParseJWT: токен успешно разобран")
	return claims, nil
}
