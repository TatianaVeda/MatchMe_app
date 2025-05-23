package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"m/backend/config"
	"m/backend/controllers"
	"m/backend/middleware"
	"m/backend/models"
	"m/backend/routes"
	"m/backend/services"
	"m/backend/sockets"

	"github.com/go-redis/redis/v8"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func setupLogger() {
	switch config.AppConfig.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {

	depsFlag := flag.Bool("deps", false, "Install dependencies")
	flag.Parse()

	if *depsFlag {
		installDependencies()
		return
	}
	_ = godotenv.Load("config/config_local.env")

	config.LoadConfig()

	setupLogger()
	log.Infof("Уровень логирования: %s", config.AppConfig.LogLevel)

	db, err := models.InitDB(config.AppConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.RedisURL,
		Password: "",
		DB:       0,
	})
	presenceService := services.NewPresenceService(rdb)

	sockets.SetDB(db)
	controllers.InitChatsController(db, presenceService)
	controllers.InitRecommendationControllerService(db, presenceService)
	router := mux.NewRouter()
	router.Use(middleware.CorsMiddleware)

	routes.InitRoutes(router, db, presenceService)

	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	go func() {
		wsAddr := ":" + config.AppConfig.WebSocketPort
		if err := sockets.InitWebSocketServer(presenceService, wsAddr); err != nil {
			log.Fatalf("Ошибка запуска WebSocket сервера: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:         ":" + config.AppConfig.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infof("Сервер запущен на порту %s", config.AppConfig.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка при запуске сервера: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("Сервер завершает работу...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении работы сервера: %v", err)
	}

	log.Info("Сервер успешно завершил работу")
}

func installDependencies() {
	deps := []string{
		"github.com/golang-jwt/jwt",
		"golang.org/x/crypto/bcrypt",
		"github.com/lib/pq",
		"github.com/gorilla/websocket",
		"github.com/google/uuid",
		"gorm.io/driver/postgres",
		"gorm.io/gorm",
		"github.com/golang-jwt/jwt/v4",
		"github.com/joho/godotenv",
		"github.com/sirupsen/logrus",
		"github.com/go-redis/redis/v8@latest",
	}

	for _, dep := range deps {
		fmt.Println("Installing", dep)
		cmd := exec.Command("go", "get", dep)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error installing %s: %v\n", dep, err)
		}
	}

	fmt.Println("Running 'go mod tidy'")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running 'go mod tidy': %v\n", err)
	}
}
