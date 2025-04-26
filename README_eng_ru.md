# Match Me - Meeting Application

> A modern meeting application with Go backend, React frontend and real-time communication via WebSockets.

<p align="center">
  <a href="#english">English</a> |
  <a href="#russian">Русский</a>
</p>

<a id="english"></a>
## Table of Contents

- [Technologies](#technologies)
- [Requirements](#requirements)
- [Installation and Setup](#installation-and-setup)
- [Project Structure](#project-structure)
- [API and Ports](#api-and-ports)
- [Development](#development)
- [Features](#features)

## Technologies

### Backend:
- **Go** - main development language
- **Gorilla Mux** - HTTP router
- **Gorilla WebSocket** - WebSocket support
- **GORM** - ORM for database operations
- **JWT** - user authentication
- **PostgreSQL** - data storage

### Frontend:
- **React** - UI library
- **React Router** - routing
- **Material UI** - interface components
- **Formik** and **Yup** - form management and validation
- **Axios** - HTTP client
- **React Toastify** - notifications

### Infrastructure:
- **Docker** and **Docker Compose** - containerization
- **Concurrently** - parallel service execution

## Requirements

- **Node.js** (v16+)
- **Go** (v1.19+)
- **Docker** and **Docker Compose**
- **WSL2** (for Windows) with Docker Desktop integration enabled

## Installation and Setup

### Automatic Installation (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/match-me.git
   cd match-me
   ```

2. Make the setup script executable:
   ```bash
   chmod +x setup.sh
   ```

3. Run the setup script:
   ```bash
   ./setup.sh
   ```
   The script will automatically configure everything needed: install Docker (if missing), create PostgreSQL database, install dependencies for both backend and frontend.

4. After installation is complete, start the application:
   ```bash
   npm run dev
   ```

5. Open the application in your browser:
   - Frontend: [http://localhost:3000](http://localhost:3000)
   - Backend API: [http://localhost:8080](http://localhost:8080)

### Manual Installation

If automatic installation doesn't suit your needs, perform these steps manually:

1. Install Docker and Docker Compose
2. Start PostgreSQL:
   ```bash
   docker compose up -d db
   ```
3. Install backend dependencies:
   ```bash
   cd backend
   go mod tidy
   go run main.go -deps
   cd ..
   ```
4. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   cd ..
   ```
5. Start the application:
   ```bash
   npm run dev
   ```

## Project Structure

```
match-me/
├── backend/              # Go server
│   ├── config/           # Configuration
│   ├── controllers/      # API controllers
│   ├── db/               # Database settings
│   ├── middleware/       # Middleware (auth, etc.)
│   ├── models/           # Data models
│   ├── routes/           # API routes
│   ├── services/         # Business logic
│   ├── sockets/          # WebSocket handlers
│   ├── static/           # Static files (images)
│   ├── tests/            # Tests
│   ├── utils/            # Helper functions
│   ├── go.mod            # Go dependencies
│   └── main.go           # Entry point
├── frontend/             # React application
│   ├── public/           # Static files
│   ├── src/              # React source code
│   └── package.json      # NPM dependencies
├── docker-compose.yml    # Docker configuration
├── package.json          # Root NPM scripts
└── setup.sh              # Setup script
```

## API and Ports

- **3000** - Frontend (React)
- **8080** - Backend API (Go)
- **8081** - WebSocket server
- **5433** - PostgreSQL

## Development

### Available Scripts

```bash
# Full installation
npm run all

# Run the entire application
npm run dev

# Database only
npm run dev:db

# Backend only
npm run dev:backend

# Frontend only
npm run dev:frontend
```

### Working with the Database

To access PostgreSQL:

```bash
docker exec -it m_postgres psql -U user -d sopostavmenya
```

## Features

- **JWT Authentication**: Secure user authentication
- **WebSockets**: Real-time communication
- **Image Upload**: Support for avatars and media content
- **Data Validation**: Strict validation of input data
- **Docker**: Easy deployment in any environment
- **CORS**: Configured security for cross-domain requests

---

<a id="russian"></a>
# Match Me - Приложение для знакомств

> Современное приложение для знакомств с бэкендом на Go, фронтендом на React и поддержкой общения в реальном времени через WebSockets.

## Содержание

- [Технологии](#технологии)
- [Требования](#требования)
- [Установка и запуск](#установка-и-запуск)
- [Структура проекта](#структура-проекта)
- [API и порты](#api-и-порты)
- [Разработка](#разработка)
- [Особенности](#особенности)

## Технологии

### Backend:
- **Go** - основной язык разработки
- **Gorilla Mux** - HTTP маршрутизатор
- **Gorilla WebSocket** - поддержка WebSocket
- **GORM** - ORM для работы с базой данных
- **JWT** - аутентификация пользователей
- **PostgreSQL** - хранение данных

### Frontend:
- **React** - UI библиотека
- **React Router** - маршрутизация
- **Material UI** - компоненты интерфейса
- **Formik** и **Yup** - управление формами и валидация
- **Axios** - HTTP клиент
- **React Toastify** - уведомления

### Инфраструктура:
- **Docker** и **Docker Compose** - контейнеризация
- **Concurrently** - параллельный запуск сервисов

## Требования

- **Node.js** (v16+)
- **Go** (v1.19+)
- **Docker** и **Docker Compose**
- **WSL2** (для Windows) с включенной интеграцией Docker Desktop

## Установка и запуск

### Автоматическая установка (рекомендуется)

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/your-username/match-me.git
   cd match-me
   ```

2. Сделайте скрипт установки исполняемым:
   ```bash
   chmod +x setup.sh
   ```

3. Запустите скрипт установки:
   ```bash
   ./setup.sh
   ```
   Скрипт автоматически настроит все необходимое: установит Docker (если отсутствует), создаст базу данных PostgreSQL, установит зависимости для backend и frontend.

4. После завершения установки запустите приложение:
   ```bash
   npm run dev
   ```

5. Откройте приложение в браузере:
   - Frontend: [http://localhost:3000](http://localhost:3000)
   - Backend API: [http://localhost:8080](http://localhost:8080)

### Ручная установка

Если автоматическая установка не подходит, выполните шаги вручную:

1. Установите Docker и Docker Compose
2. Запустите PostgreSQL:
   ```bash
   docker compose up -d db
   ```
3. Установите зависимости для backend:
   ```bash
   cd backend
   go mod tidy
   go run main.go -deps
   cd ..
   ```
4. Установите зависимости для frontend:
   ```bash
   cd frontend
   npm install
   cd ..
   ```
5. Запустите приложение:
   ```bash
   npm run dev
   ```

## Структура проекта

```
match-me/
├── backend/              # Go сервер
│   ├── config/           # Конфигурация
│   ├── controllers/      # Контроллеры API
│   ├── db/               # Настройки базы данных
│   ├── middleware/       # Промежуточное ПО (авторизация и т.д.)
│   ├── models/           # Модели данных
│   ├── routes/           # Маршруты API
│   ├── services/         # Бизнес-логика
│   ├── sockets/          # WebSocket обработчики
│   ├── static/           # Статические файлы (изображения)
│   ├── tests/            # Тесты
│   ├── utils/            # Вспомогательные функции
│   ├── go.mod            # Go зависимости
│   └── main.go           # Точка входа
├── frontend/             # React приложение
│   ├── public/           # Статические файлы
│   ├── src/              # Исходный код React
│   └── package.json      # NPM зависимости
├── docker-compose.yml    # Конфигурация Docker
├── package.json          # Корневые NPM скрипты
└── setup.sh              # Скрипт установки
```

## API и порты

- **3000** - Frontend (React)
- **8080** - Backend API (Go)
- **8081** - WebSocket сервер
- **5433** - PostgreSQL

## Разработка

### Доступные скрипты

```bash
# Полная установка
npm run all

# Запуск всего приложения
npm run dev

# Только база данных
npm run dev:db

# Только backend
npm run dev:backend

# Только frontend
npm run dev:frontend
```

### Работа с базой данных

Для доступа к PostgreSQL:

```bash
docker exec -it m_postgres psql -U user -d sopostavmenya
```

## Особенности

- **JWT аутентификация**: Безопасная аутентификация пользователей
- **WebSockets**: Общение в реальном времени
- **Загрузка изображений**: Поддержка аватаров и медиа-контента
- **Валидация данных**: Строгая проверка вводимых данных
- **Docker**: Простое развертывание в любой среде
- **CORS**: Настроена безопасность для междоменных запросов
