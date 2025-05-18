package routes

import (
	"m/backend/controllers"
	"m/backend/middleware"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitRoutes инициализирует все маршруты приложения.
func InitRoutes(router *mux.Router, db *gorm.DB) {
	logrus.Info("Initializing routes...")
	// Инициализируем контроллеры с подключением к базе данных.
	controllers.InitUserController(db)
	// Инициализируем контроллер для рекомендаций через сервисный слой.
	controllers.InitRecommendationControllerService(db)
	controllers.InitConnectionsController(db)
	controllers.InitChatsController(db)
	controllers.InitProfileController(db)        // Инициализация контроллеров профиля
	controllers.InitFixturesController(db)       // Инициализация фикстур
	controllers.InitAuthenticationController(db) // Добавляем инициализацию нашего нового контроллера
	controllers.InitPreferencesController(db)
	controllers.InitCitiesController(db)

	// --- Публичные эндпоинты пользователей ---

	router.HandleFunc("/signup", controllers.Signup).Methods(http.MethodPost)
	//router.HandleFunc("/users/{id}", controllers.GetUser).Methods(http.MethodGet)
	//router.HandleFunc("/users/{id}/profile", controllers.GetUserProfile).Methods(http.MethodGet)
	//router.HandleFunc("/users/{id}/bio", controllers.GetUserBio).Methods(http.MethodGet)
	// Эндпоинт для обновления токенов не требует аутентификации (он принимает refresh токен)
	router.HandleFunc("/refresh", controllers.RefreshToken).Methods(http.MethodPost)
	router.HandleFunc("/login", controllers.Login).Methods(http.MethodPost)
	router.HandleFunc("/cities", controllers.GetCities).Methods(http.MethodGet)

	// --- Эндпоинты для аутентифицированного пользователя ---
	// Создаем subrouter для защищенных маршрутов и подключаем AuthMiddleware.
	authRouter := router.PathPrefix("/").Subrouter()
	authRouter.Use(controllers.AuthMiddleware) // Подключаем middleware аутентификации

	//test
	// authRouter.HandleFunc("/{anything}", func(w http.ResponseWriter, r *http.Request) {
	// 	logrus.Infof("REQUEST Request path: %s", r.URL.Path)
	// }).Methods(http.MethodGet)

	// usersRouter := authRouter.PathPrefix("/users").Subrouter()
	// usersRouter.HandleFunc("/{id}", controllers.GetUser).Methods("GET")
	// usersRouter.HandleFunc("/{id}/bio", controllers.GetUserBio).Methods("GET")
	// usersRouter.HandleFunc("/{id}/profile", controllers.GetUserProfile).Methods("GET")

	//moved to protected routes
	authRouter.HandleFunc("/users/{id}", controllers.GetUser).Methods(http.MethodGet)
	authRouter.HandleFunc("/users/{id}/profile", controllers.GetUserProfile).Methods(http.MethodGet)
	authRouter.HandleFunc("/users/{id}/bio", controllers.GetUserBio).Methods(http.MethodGet)
	authRouter.HandleFunc("/users/{id}/profile", controllers.GetUserProfile).Methods(http.MethodGet)

	authRouter.HandleFunc("/me", controllers.GetCurrentUser).Methods(http.MethodGet)
	authRouter.HandleFunc("/me/profile", controllers.GetCurrentUserProfile).Methods(http.MethodGet)
	authRouter.HandleFunc("/me/bio", controllers.GetCurrentUserBio).Methods(http.MethodGet)
	// Обновление профиля и биографии через PUT-метод.
	authRouter.HandleFunc("/me/profile", controllers.UpdateCurrentUserProfile).Methods(http.MethodPut)
	authRouter.HandleFunc("/me/bio", controllers.UpdateCurrentUserBio).Methods(http.MethodPut)
	authRouter.HandleFunc("/me/location", controllers.UpdateCurrentUserLocation).Methods(http.MethodPut)
	// Загрузка фотографии профиля.
	authRouter.HandleFunc("/me/photo", controllers.UploadUserPhoto).Methods(http.MethodPost)
	authRouter.HandleFunc("/me/photo", controllers.DeleteUserPhoto).Methods(http.MethodDelete)
	authRouter.HandleFunc("/logout", controllers.Logout).Methods(http.MethodPost)
	authRouter.HandleFunc("/me/email", controllers.UpdateEmail).Methods(http.MethodPut)
	authRouter.HandleFunc("/me/password", controllers.UpdatePassword).Methods(http.MethodPut)

	// --- Остальные эндпоинты (рекомендации, связи, чат) ---
	authRouter.HandleFunc("/recommendations", controllers.GetRecommendations).Methods(http.MethodGet)
	authRouter.HandleFunc("/recommendations/{id}/decline", controllers.DeclineRecommendation).Methods(http.MethodPost)
	authRouter.HandleFunc("/connections", controllers.GetConnections).Methods(http.MethodGet)
	authRouter.HandleFunc("/connections/pending", controllers.GetPendingConnections).Methods(http.MethodGet)
	authRouter.HandleFunc("/connections/{id}", controllers.PostConnection).Methods(http.MethodPost)
	authRouter.HandleFunc("/connections/{id}", controllers.PutConnection).Methods(http.MethodPut)
	authRouter.HandleFunc("/connections/{id}", controllers.DeleteConnection).Methods(http.MethodDelete)
	authRouter.HandleFunc("/chats", controllers.CreateOrGetChat).Methods(http.MethodPost)
	authRouter.HandleFunc("/chats", controllers.GetChats).Methods(http.MethodGet)
	authRouter.HandleFunc("/chats/{chatId}", controllers.GetChatHistory).Methods(http.MethodGet)
	authRouter.HandleFunc("/chats/{chatId}/messages", controllers.PostMessage).Methods(http.MethodPost)
	authRouter.HandleFunc("/me/preferences", controllers.GetPreferences).Methods(http.MethodGet)
	authRouter.HandleFunc("/me/preferences", controllers.UpdatePreferences).Methods(http.MethodPut)

	// --- Административные эндпоинты ---
	// создаём отдельный subrouter с префиксом `/admin`, внутри — AuthMiddleware и затем AdminOnly
	adminOnly := middleware.AdminOnly(db)
	adminRouter := authRouter.PathPrefix("/admin").Subrouter()
	// здесь уже применяется AuthMiddleware от authRouter, добавляем AdminOnly
	adminRouter.Use(adminOnly)

	// фикстуры: сброс и генерация пользователей
	adminRouter.HandleFunc("/reset-fixtures", controllers.ResetFixtures).Methods(http.MethodPost)
	adminRouter.HandleFunc("/generate-fixtures", controllers.GenerateFixtures).Methods(http.MethodPost)

	logrus.Info("Routes successfully initialized")
}
