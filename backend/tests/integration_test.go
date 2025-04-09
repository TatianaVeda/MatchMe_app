package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"m/backend/controllers"
	"m/backend/models"
	"m/backend/routes"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// setupRouter создаёт роутер с подключёнными маршрутами.
func setupRouter(db *gorm.DB) *mux.Router {
	r := mux.NewRouter()
	routes.InitRoutes(r, db)
	return r
}

// createTestJWT создаёт JWT-токен для теста.
func createTestJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("supersecret"))
}

func TestGetCurrentUserIntegration(t *testing.T) {
	db := SetupTestDB(t)
	// Создаём тестового пользователя.
	user := models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "dummyHash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user.Profile = models.Profile{
		UserID:    user.ID,
		FirstName: "Test",
		LastName:  "User",
		PhotoURL:  "/static/images/default.png",
	}
	db.Create(&user)

	controllers.InitUserController(db)
	r := setupRouter(db)

	token, err := createTestJWT(user.ID.String())
	if err != nil {
		t.Fatalf("Ошибка создания JWT: %v", err)
	}

	req, _ := http.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получен %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Ошибка разбора ответа: %v", err)
	}

	if resp["email"] != user.Email {
		t.Errorf("Ожидался email %s, получен %v", user.Email, resp["email"])
	}
}
