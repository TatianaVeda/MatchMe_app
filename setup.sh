#!/bin/bash

# Проверка наличия Docker Compose
if ! command -v docker-compose &> /dev/null; then
  echo "Docker Compose не найден. Установите его и повторите попытку."
  exit 1
fi

echo "Запускаем PostgreSQL через Docker Compose..."
docker-compose up -d

# Проверка доступности PostgreSQL
echo "Ждём 5 секунд, пока PostgreSQL полностью запустится..."
sleep 5
docker exec m_postgres pg_isready
if [ $? -ne 0 ]; then
  echo "PostgreSQL не готов! Проверьте логи контейнера:"
  docker logs m_postgres
  exit 1
fi
echo "PostgreSQL запущен и готов к подключению."

# Установка зависимостей для backend
echo "Переходим в папку backend и устанавливаем зависимости Go..."
cd backend || { echo "Папка backend не найдена!"; exit 1; }

# Если файл go.mod отсутствует, инициализируем модуль
if [ ! -f "go.mod" ]; then
  echo "Файл go.mod не найден. Инициализируем модуль..."
  go mod init m/backend
fi

# Обновление/создание файла go.sum через go mod tidy (подсчитываются и записываются контрольные суммы)
echo "Обновляем зависимости (go.sum)..."
go mod sum

# Запуск установки дополнительных зависимостей (если ваш код обрабатывает флаг -deps)
echo "Запускаем установку зависимостей с помощью 'go run main.go -deps'..."
go run main.go -deps

# Возвращаемся в корневую директорию проекта
cd ..

# Установка зависимостей для frontend
echo "Переходим в папку frontend и устанавливаем зависимости Node.js..."
cd frontend || { echo "Папка frontend не найдена!"; exit 1; }
npm install

# Возвращаемся в корневую директорию проекта
cd ..

echo "Окружение успешно настроено!"
echo "-------------------------------------------------"
echo "Для запуска приложения используйте команду: npm run dev"
echo "-------------------------------------------------"
