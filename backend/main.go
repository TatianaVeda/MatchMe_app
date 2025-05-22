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

	// Определяем флаг для установки зависимостей
	depsFlag := flag.Bool("deps", false, "Install dependencies")
	flag.Parse()

	if *depsFlag {
		installDependencies()
		return
	}
	// Загружаем переменные окружения из файла конфигурации
	_ = godotenv.Load("config/config_local.env")

	// Загружаем конфигурацию приложения
	config.LoadConfig()

	// Настраиваем логирование в соответствии с параметром LOG_LEVEL
	setupLogger()
	log.Infof("Уровень логирования: %s", config.AppConfig.LogLevel)

	// Инициализируем подключение к базе данных
	db, err := models.InitDB(config.AppConfig.DatabaseURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	// 1) Настраиваем Redis для presence
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.RedisURL,
		Password: "",
		DB:       0,
	})
	presenceService := services.NewPresenceService(rdb)

	// Если используется GORM v2, явного закрытия соединения не требуется

	// Устанавливаем экземпляр базы данных для пакета sockets,
	// чтобы функции обновления онлайн-статуса могли её использовать.
	sockets.SetDB(db)
	// Передаём PresenceService в контроллер чатов и в контроллер рекомендаций
	controllers.InitChatsController(db, presenceService)
	controllers.InitRecommendationControllerService(db, presenceService)
	// Создаем маршрутизатор и применяем глобальный CORS-мидлвар
	router := mux.NewRouter()
	router.Use(middleware.CorsMiddleware)

	// Инициализируем все маршруты, передавая подключение к БД
	//routes.InitRoutes(router, db)
	routes.InitRoutes(router, db, presenceService)

	// Раздача статики для картинок
	// любой запрос к /static/... будет браться из папки ./static
	// router.PathPrefix("/static/").Handler(
	// 	http.StripPrefix("/static/",
	// 		http.FileServer(http.Dir(config.AppConfig.MediaUploadDir)),
	// 	),
	// )

	router.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Запускаем WebSocket-сервер в отдельной горутине (на порту, указанном в конфигурации)
	// go func() {
	// 	// sockets.InitWebSocketServer(presenceService)
	// 	// wsAddr := ":" + config.AppConfig.WebSocketPort

	// 	wsAddr := ":" + config.AppConfig.WebSocketPort
	// 	err := sockets.InitWebSocketServer(presenceService, wsAddr)
	// 	if err := sockets.RunWebSocketServer(wsAddr); err != nil {
	// 		log.Fatalf("Ошибка запуска WebSocket сервера: %v", err)
	// 	}

	// 	// if err := sockets.RunWebSocketServer(wsAddr); err != nil {
	// 	// 	log.Fatalf("Ошибка запуска WebSocket сервера: %v", err)
	// 	// }
	// }()

	go func() {
		wsAddr := ":" + config.AppConfig.WebSocketPort
		if err := sockets.InitWebSocketServer(presenceService, wsAddr); err != nil {
			log.Fatalf("Ошибка запуска WebSocket сервера: %v", err)
		}
	}()

	// Создаем HTTP-сервер с настройками таймаутов для graceful shutdown
	srv := &http.Server{
		Addr:         ":" + config.AppConfig.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем HTTP-сервер в отдельной горутине
	go func() {
		log.Infof("Сервер запущен на порту %s", config.AppConfig.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка при запуске сервера: %v", err)
		}
	}()

	// Ожидаем сигнала прерывания (например, Ctrl+C) для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("Сервер завершает работу...")

	// Создаем контекст с таймаутом для корректного завершения работы сервера
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
