package routes

import (
	"m/backend/controllers"
	"m/backend/middleware"
	"m/backend/services"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitRoutes initializes all application routes, connects controllers, middleware, and services.
// Uses mux.Router, GORM, and services for users, chats, recommendations, etc.
func InitRoutes(router *mux.Router, db *gorm.DB, ps *services.PresenceService) {
	logrus.Info("Initializing routes...")
	controllers.InitUserController(db)
	controllers.InitConnectionsController(db)

	controllers.InitRecommendationControllerService(db, ps)
	controllers.InitChatsController(db, ps)
	controllers.InitProfileController(db)
	controllers.InitFixturesController(db)
	controllers.InitAuthenticationController(db)
	controllers.InitPreferencesController(db)
	controllers.InitCitiesController(db)
	presenceCtrl := controllers.NewPresenceController(ps)

	router.HandleFunc("/signup", controllers.Signup).Methods(http.MethodPost)
	router.HandleFunc("/refresh", controllers.RefreshToken).Methods(http.MethodPost)
	router.HandleFunc("/login", controllers.Login).Methods(http.MethodPost)
	router.HandleFunc("/cities", controllers.GetCities).Methods(http.MethodGet)

	router.HandleFunc("/api/user/online", presenceCtrl.GetOnlineStatus).Methods("GET")
	router.HandleFunc("/api/user/online/batch", presenceCtrl.GetMultipleOnlineStatus).Methods("GET")

	authRouter := router.PathPrefix("/").Subrouter()
	authRouter.Use(controllers.AuthMiddleware)

	authRouter.HandleFunc("/users/{id}", controllers.GetUser).Methods(http.MethodGet)
	authRouter.HandleFunc("/users/{id}/bio", controllers.GetUserBio).Methods(http.MethodGet)
	authRouter.HandleFunc("/users/{id}/profile", controllers.GetUserProfile).Methods(http.MethodGet)

	authRouter.HandleFunc("/me", controllers.GetCurrentUser).Methods(http.MethodGet)
	authRouter.HandleFunc("/me/profile", controllers.GetCurrentUserProfile).Methods(http.MethodGet)
	authRouter.HandleFunc("/me/bio", controllers.GetCurrentUserBio).Methods(http.MethodGet)

	authRouter.HandleFunc("/me/profile", controllers.UpdateCurrentUserProfile).Methods(http.MethodPut)
	authRouter.HandleFunc("/me/bio", controllers.UpdateCurrentUserBio).Methods(http.MethodPut)
	authRouter.HandleFunc("/me/location", controllers.UpdateCurrentUserLocation).Methods(http.MethodPut)

	authRouter.HandleFunc("/me/photo", controllers.UploadUserPhoto).Methods(http.MethodPost)
	authRouter.HandleFunc("/me/photo", controllers.DeleteUserPhoto).Methods(http.MethodDelete)
	authRouter.HandleFunc("/logout", controllers.Logout).Methods(http.MethodPost)
	authRouter.HandleFunc("/me/email", controllers.UpdateEmail).Methods(http.MethodPut)
	authRouter.HandleFunc("/me/password", controllers.UpdatePassword).Methods(http.MethodPut)

	authRouter.HandleFunc("/recommendations", controllers.GetRecommendations).Methods(http.MethodGet)
	authRouter.HandleFunc("/recommendations/{id}/decline", controllers.DeclineRecommendation).Methods(http.MethodPost)
	authRouter.HandleFunc("/connections", controllers.GetConnections).Methods(http.MethodGet)
	authRouter.HandleFunc("/connections/pending", controllers.GetPendingConnections).Methods(http.MethodGet)
	authRouter.HandleFunc("/connections/sent", controllers.GetSentConnections).Methods(http.MethodGet)
	authRouter.HandleFunc("/connections/{id}", controllers.PostConnection).Methods(http.MethodPost)
	authRouter.HandleFunc("/connections/{id}", controllers.PutConnection).Methods(http.MethodPut)
	authRouter.HandleFunc("/connections/{id}", controllers.DeleteConnection).Methods(http.MethodDelete)
	authRouter.HandleFunc("/chats", controllers.CreateOrGetChat).Methods(http.MethodPost)
	authRouter.HandleFunc("/chats", controllers.GetChats).Methods(http.MethodGet)
	authRouter.HandleFunc("/chats/{chatId}", controllers.GetChatHistory).Methods(http.MethodGet)
	authRouter.HandleFunc("/chats/{chatId}/messages", controllers.PostMessage).Methods(http.MethodPost)
	authRouter.HandleFunc("/me/preferences", controllers.GetPreferences).Methods(http.MethodGet)
	authRouter.HandleFunc("/me/preferences", controllers.UpdatePreferences).Methods(http.MethodPut)
	adminOnly := middleware.AdminOnly(db)
	adminRouter := authRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(adminOnly)

	adminRouter.HandleFunc("/reset-fixtures", controllers.ResetFixtures).Methods(http.MethodPost)
	adminRouter.HandleFunc("/generate-fixtures", controllers.GenerateFixtures).Methods(http.MethodPost)

	logrus.Info("Routes successfully initialized")
}
