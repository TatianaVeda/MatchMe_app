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
- [Setting Recommendation Radius](#setting-recommendation-radius)

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

### Environment Configuration

Before running the application, ensure that the `config_local.env` file is properly configured. This file contains essential environment variables such as:

- **SERVER_PORT**: The port on which the backend server will run.
- **WEBSOCKET_PORT**: The port for WebSocket connections.
- **DATABASE_URL**: Connection string for the PostgreSQL database.
- **JWT_SECRET**: Secret key for signing JWT tokens.
- **JWT_EXPIRES_IN**: Expiration time for JWT tokens.
- **MEDIA_UPLOAD_DIR**: Directory for storing uploaded media files.
- **ENVIRONMENT**: The environment mode (e.g., development, production).
- **ALLOWED_ORIGINS**: Comma-separated list of allowed origins for CORS.
- **LOG_LEVEL**: Logging level (e.g., debug, info, warn, error).

Ensure these variables are updated according to your local setup.

### Generating Dummy Users

To generate dummy users for testing purposes, you can use the `ResetFixtures` 
function in the `fixtures.go` file. This function will reset the database and 
populate it with a specified number of dummy users. You can trigger this function 
via an API call to the backend server. Follow these steps:

1. **Start the Backend Server**: Ensure your backend server is running. You can start it using:
   ```bash
   npm run dev:backend
   ```

2. **Trigger the ResetFixtures Function**: Use a tool like `curl` or Postman to send a request to the backend API to reset the database and generate dummy users. For example, using `curl`:
   ```bash
   curl -X POST http://localhost:8080/api/reset-fixtures
   ```

The `createAdminUser` function in the `create_admin_user.go` file is used to create 
an initial admin user in the database. This is typically done during the initial 
setup to ensure that an admin account is available for managing the application.This will reset the database and populate it with dummy users as defined in the `fixtures.go` file.

### Automatic Installation (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/ihorshaposhnik/m.git
   cd m
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

### Docker Installation

For setting up Docker on different operating systems, follow these steps:

1. **Windows**:
   - Install Docker Desktop from the [official website](https://www.docker.com/products/docker-desktop).
   - Enable WSL2 integration in Docker Desktop settings.
   - Start Docker Desktop to use Docker in WSL2.

2. **macOS**:
   - Install Docker Desktop for Mac from the [official website](https://www.docker.com/products/docker-desktop).
   - Start Docker Desktop to use Docker in the terminal.

3. **Linux**:
   - Install Docker using your distribution's package manager (e.g., `apt` for Ubuntu).
   - Ensure Docker is running as a daemon.
   - Use Docker commands in the terminal as usual.

## Project Structure

```
match-me/m
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

### Running create_admin_user.go

To create an admin account, follow these steps:

1. **Run the create_admin_user.go file**:
   - Navigate to the `backend` directory and execute the command:
     ```bash
     go run create_admin_user.go
     ```
   - This will create an admin account in the database.

Ensure the database is configured and accessible before running this command.

### Using create_admin_user.go

The `create_admin_user.go` file allows you to perform the following actions:

- **Reset the database:**
  - Use the `-resetDB` flag to reset the database.
  - Example command:
    ```bash
    go run create_admin_user.go -resetDB
    ```

- **Create an admin user:**
  - Use the `-createAdmin` flag to create an admin user account.
  - Example command:
    ```bash
    go run create_admin_user.go -createAdmin
    ```

These commands allow you to manage the database and admin accounts without modifying the main application code.

### Creating Admin Profile via Browser

You can also create an admin profile through the browser and then obtain a JWT token using the login endpoint.

### Obtaining JWT Token

To obtain a JWT token, follow these steps:

1. **Send a request to the `/login` endpoint**: Use `curl` to send a request with your credentials.

   ```bash
   curl -X POST http://localhost:8080/login \
   -H "Content-Type: application/json" \
   -d '{"email":"your_admin_email","password":"your_password"}'
   ```

   Replace `your_admin_email` and `your_password` with your actual credentials.

This allows you to manage the application using the admin profile created through the browser.

## Setting Recommendation Radius

To search for recommendations within a specific geographical radius, navigate to the "Settings" page of the application. At the bottom of the settings page, you will find an option to set the maximum radius for recommendations. Adjusting this value enables proximity-based filtering for recommendations.

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
- [Настройка радиуса для рекомендаций](#настройка-радиуса-для-рекомендаций)

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

### Конфигурация окружения

Перед запуском приложения убедитесь, что файл `config_local.env` правильно настроен. Этот файл содержит важные переменные окружения, такие как:

- **SERVER_PORT**: Порт, на котором будет работать сервер.
- **WEBSOCKET_PORT**: Порт для WebSocket соединений.
- **DATABASE_URL**: Строка подключения к базе данных PostgreSQL.
- **JWT_SECRET**: Секретный ключ для подписи JWT токенов.
- **JWT_EXPIRES_IN**: Время истечения JWT токенов.
- **MEDIA_UPLOAD_DIR**: Директория для хранения загружаемых медиафайлов.
- **ENVIRONMENT**: Режим работы приложения (например, development, production).
- **ALLOWED_ORIGINS**: Список разрешенных источников для CORS.
- **LOG_LEVEL**: Уровень логирования (например, debug, info, warn, error).

Убедитесь, что эти переменные обновлены в соответствии с вашей локальной конфигурацией.

### Генерация фиктивных пользователей

Для генерации фиктивных пользователей в целях тестирования вы можете использовать 
функцию `ResetFixtures` в файле `fixtures.go`. Эта функция сбросит базу данных и 
заполнит её указанным количеством фиктивных пользователей. Вы можете вызвать эту 
функцию через API-запрос к серверу. В целях тестирования выполните следующие шаги:

1. **Запустите сервер**: Убедитесь, что ваш сервер запущен. Вы можете запустить его с помощью команды:
   ```bash
   npm run dev:backend
   ```

2. **Вызовите функцию ResetFixtures**: Используйте инструмент, такой как `curl` или Postman, чтобы отправить запрос к API сервера для сброса базы данных и генерации фиктивных пользователей. Например, с помощью `curl`:
   ```bash
   curl -X POST http://localhost:8080/api/reset-fixtures
   ```

Функция `createAdminUser` в файле `create_admin_user.go` используется для создания 
начального администратора в базе данных. Это обычно делается во время начальной 
настройки, чтобы обеспечить наличие учетной записи администратора для управления 
приложением. Это сбросит базу данных и заполнит её фиктивными пользователями, как определено в файле `fixtures.go`.

### Automatic Installation (Recommended)

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/ihorshaposhnik/m.git
   git
   cd m
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

### Docker Installation for Different Operating Systems

1. **Windows**:
   - Установите Docker Desktop с [официального сайта](https://www.docker.com/products/docker-desktop).
   - Включите интеграцию с WSL2 в настройках Docker Desktop.
   - Запустите Docker Desktop, чтобы использовать Docker в WSL2.

2. **macOS**:
   - Установите Docker Desktop для Mac с [официального сайта](https://www.docker.com/products/docker-desktop).
   - Запустите Docker Desktop, чтобы использовать Docker в терминале.

3. **Linux**:
   - Установите Docker через пакетный менеджер вашей дистрибуции (например, `apt` для Ubuntu).
   - Убедитесь, что Docker запущен как демон.
   - Используйте команды Docker в терминале, как обычно.

## Структура проекта

```
match-me/m
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

### Запуск создания администратора

Для создания учетной записи администратора выполните следующий шаг:

1. **Запустите файл create_admin_user.go**: 
   - Перейдите в директорию `backend` и выполните команду:
     ```bash
     go run create_admin_user.go
     ```
   - Это создаст учетную запись администратора в базе данных.

Убедитесь, что база данных настроена и доступна перед запуском этой команды.

### Использование create_admin_user.go

Файл `create_admin_user.go` позволяет выполнять следующие действия:

- **Сброс базы данных:**
  - Используйте флаг `-resetDB` для сброса базы данных.
  - Пример команды:
    ```bash
    go run create_admin_user.go -resetDB
    ```

- **Создание администратора:**
  - Используйте флаг `-createAdmin` для создания учетной записи администратора.
  - Пример команды:
    ```bash
    go run create_admin_user.go -createAdmin
    ```

Эти команды позволяют управлять базой данных и учетными записями администратора без изменения основного кода приложения.

### Создание профиля администратора через браузер

Вы также можете создать профиль администратора через браузер, а затем получить JWT токен, используя эндпоинт для входа в систему.

### Получение JWT токена

Чтобы получить JWT токен, выполните следующие шаги:

1. **Отправьте запрос на эндпоинт `/login`**: Используйте `curl` для отправки запроса с вашими учетными данными.

   ```bash
   curl -X POST http://localhost:8080/login \
   -H "Content-Type: application/json" \
   -d '{"email":"your_admin_email","password":"your_password"}'
   ```

   Замените `your_admin_email` и `your_password` на ваши фактические данные.

Это позволяет управлять приложением, используя профиль администратора, созданный через браузер.

## Настройка радиуса для рекомендаций

Для поиска рекомендаций в определённом географическом радиусе перейдите на страницу "Настройки" приложения. Внизу страницы настроек вы найдёте опцию для установки максимального радиуса рекомендаций. Настройка этого значения позволяет использовать фильтрацию по близости для рекомендаций.
