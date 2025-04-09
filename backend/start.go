// package main

// import (
// 	"fmt"
// 	"os"
// 	"os/exec"
// )

// func mainStart() {
// 	// Список зависимостей, которые нужно установить.
// 	deps := []string{
// 		"github.com/golang-jwt/jwt",
// 		"golang.org/x/crypto/bcrypt",
// 		"github.com/lib/pq",
// 		"github.com/gorilla/websocket",
// 		"github.com/google/uuid",
// 		"gorm.io/driver/postgres",
// 		"gorm.io/gorm",
// 		"github.com/golang-jwt/jwt/v4",
// 		"github.com/joho/godotenv",
// 		"github.com/sirupsen/logrus",
// 	}

// 	// Установка каждой зависимости.
// 	for _, dep := range deps {
// 		fmt.Println("Installing", dep)
// 		cmd := exec.Command("go", "get", dep)
// 		cmd.Stdout = os.Stdout
// 		cmd.Stderr = os.Stderr
// 		if err := cmd.Run(); err != nil {
// 			fmt.Printf("Error installing %s: %v\n", dep, err)
// 		}
// 	}

// 	// Очистка и обновление go.mod/go.sum.
// 	fmt.Println("Running 'go mod tidy'")
// 	cmd := exec.Command("go", "mod", "tidy")
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	if err := cmd.Run(); err != nil {
// 		fmt.Printf("Error running 'go mod tidy': %v\n", err)
// 	}
// }
