package routes

import (
	"backend/handlers"
	"backend/middlewares"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes initializes all application routes
func RegisterRoutes(router *gin.Engine) {
	// Public Routes
	api := router.Group("/api")
	{
		api.POST("/login", handlers.LoginUser)
		api.POST("/register", handlers.RegisterUser)
	}

	// Protected Routes with Authentication Middleware
	authenticated := api.Group("")
	authenticated.Use(middlewares.AuthMiddleware())
	{
		registerProfileRoutes(authenticated)
		registerUserRoutes(authenticated)
		registerRecommendationRoutes(authenticated)
		registerConnectionRoutes(authenticated)
		registerChatRoutes(authenticated)
	}
}

func registerProfileRoutes(r *gin.RouterGroup) {
	profile := r.Group("/me")
	{
		profile.GET("", handlers.GetUserProfile)
		profile.GET("/profile", handlers.GetUserProfileDetails)
		profile.GET("/bio", handlers.GetUserFavorites)
		profile.GET("/email", handlers.GetUserEmail)
		profile.PUT("", handlers.UpdateUserProfile)
		profile.PUT("/avatar", handlers.UpdateUserAvatar)
		profile.DELETE("/avatar", handlers.ResetUserAvatar)
	}
}

func registerUserRoutes(r *gin.RouterGroup) {
	users := r.Group("/users/:id")
	{
		users.GET("", handlers.GetUserNameAndPic)
		users.GET("/profile", handlers.GetUserInfo)
		users.GET("/bio", handlers.GetUserFavs)
	}

	bio := r.Group("/me/bio")
	{
		bio.GET("/locations", handlers.GetLocations)
		bio.GET("/genres", handlers.GetGenres)
		bio.GET("/movies", handlers.GetMovies)
		bio.GET("/directors", handlers.GetDirectors)
		bio.GET("/actors", handlers.GetActors)
		bio.GET("/actresses", handlers.GetActresses)
	}
}

func registerRecommendationRoutes(r *gin.RouterGroup) {
	recs := r.Group("/recommendations")
	{
		recs.GET("", handlers.MatchUsers)
		recs.POST("/scores", handlers.GetRecommendations)
		recs.PUT("/radius", handlers.SetRadius)
		recs.GET("/radius", handlers.SetRadius)
		recs.POST("/dismiss", handlers.DismissUser)
	}
}

func registerConnectionRoutes(r *gin.RouterGroup) {
	conn := r.Group("/connections")
	{
		conn.POST("/request", handlers.SendConnectionRequest)
		conn.GET("/requests", handlers.GetConnectionRequests)
		conn.POST("/accept", handlers.AcceptConnectionRequest)
		conn.POST("/reject", handlers.RejectConnectionRequest)
		conn.GET("", handlers.GetConnectedUsers)
		conn.POST("/remove", handlers.RemoveConnection)
	}
}

func registerChatRoutes(r *gin.RouterGroup) {
	chat := r.Group("/chat")
	{
		chat.POST("/send", handlers.SendChatMessage)
		chat.GET("/messages", handlers.FetchChatMessages)
		chat.GET("/last-message", handlers.RetrieveLastMessageTimestamp)
		chat.POST("/mark-read", handlers.MarkMessagesRead)
		chat.GET("/unread", handlers.FetchUnreadMessageCounts)
	}
}
