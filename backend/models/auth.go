package models

import (
	"errors"
	"m/backend/config"
	"m/backend/utils"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/sirupsen/logrus"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Errorf("HashPassword: hash generation error: %v", err)
		return "", err
	}
	logrus.Debug("HashPassword: password hashed successfully")
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		logrus.Debugf("CheckPasswordHash: does not match: %v", err)
		return false
	}
	return true
}

func CreateUser(db *gorm.DB, email, password string) (*User, error) {

	if email == config.AdminEmail {

		if password != config.AdminPassword {
			logrus.Warnf("CreateUser: invalid admin password for email %s", email)
			return nil, errors.New("invalid admin credentials")
		}

		hash, err := HashPassword(password)
		if err != nil {
			return nil, err
		}

		adminUUID := uuid.MustParse(config.AdminID)
		user := &User{
			ID:           adminUUID,
			Email:        email,
			PasswordHash: hash,
			Profile:      Profile{},
			Bio:          Bio{},
			Preference:   Preference{},
		}
		if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Create(user).Error; err != nil {
			logrus.Errorf("CreateUser (admin): admin creation error: %v", err)
			return nil, err
		}
		logrus.Infof("CreateUser: admin created with ID=%s", user.ID)
		return user, nil
	}

	if err := utils.ValidateEmail(email); err != nil {
		logrus.Warnf("CreateUser: invalid email format %s: %v", email, err)
		return nil, err
	}
	if err := utils.ValidatePassword(password); err != nil {
		logrus.Warnf("CreateUser: password does not meet requirements: %v", err)
		return nil, err
	}

	var count int64
	if err := db.Model(&User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		logrus.Errorf("CreateUser: error checking email existence %s: %v", email, err)
		return nil, err
	}
	if count > 0 {
		logrus.Warn("CreateUser: user with this email is already registered")
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
		Profile:      Profile{},
		Bio:          Bio{},
		Preference:   Preference{},
	}

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Create(user).Error; err != nil {
		logrus.Errorf("CreateUser: error creating user and associations: %v", err)
		return nil, err
	}

	defaultBio := Bio{UserID: user.ID}
	if err := db.Create(&defaultBio).Error; err != nil {
		logrus.Warnf("CreateUser: failed to create default Bio: %v", err)
	}

	logrus.Infof("CreateUser: user and related records created successfully (ID=%s)", user.ID)
	return user, nil
}

func AuthenticateUser(db *gorm.DB, email, password string) (*User, error) {

	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Warnf("AuthenticateUser: user with email %s not found", email)
			return nil, ErrUserNotFound
		}
		logrus.Errorf("AuthenticateUser: error finding user %s: %v", email, err)
		return nil, err
	}

	if !CheckPasswordHash(password, user.PasswordHash) {
		logrus.Warnf("AuthenticateUser: invalid password for user %s", email)
		return nil, ErrInvalidCredentials
	}

	logrus.Infof("AuthenticateUser: user %s authenticated successfully", email)
	return &user, nil
}

type JWTClaims struct {
	UserID uuid.UUID `json:"userId"`
	jwt.RegisteredClaims
	ExpiresAt int64 `json:"exp"`
}

func GenerateJWT(userID uuid.UUID, secret string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		ExpiresAt: time.Now().Add(72 * time.Hour).Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		logrus.Errorf("GenerateJWT: error creating token for user %s: %v", userID, err)
		return "", err
	}

	logrus.Infof("GenerateJWT: token successfully created for user %s", userID)
	return tokenString, nil
}

func ParseJWT(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		logrus.Errorf("ParseJWT: error parsing token: %v", err)
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		logrus.Warn("ParseJWT: invalid claims or token is not valid")
		return nil, errors.New("invalid token")
	}

	logrus.Debug("ParseJWT: token parsed successfully")
	return claims, nil
}

func GenerateAccessToken(userID uuid.UUID, secret string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(userID uuid.UUID, secret string, expiresInMinutes int) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiresInMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
