#!/bin/bash

# Функция для вывода сообщения об ошибке и завершения скрипта
function error_exit {
  echo "$1" >&2
  exit 1
}

npm install concurrently --save-dev

# Проверка наличия Docker Compose
if ! command -v docker-compose &> /dev/null; then
  echo "Docker Compose не найден. Пытаемся установить его..."
  # Для Ubuntu/Debian:
  sudo apt-get update
  sudo apt-get install -y docker-compose
  # Проверяем снова, если не установилось, завершаем скрипт
  if ! command -v docker-compose &> /dev/null; then
    error_exit "Не удалось установить Docker Compose. Установите его вручную и повторите попытку."
  fi
fi


echo "Запускаем PostgreSQL через Docker Compose..."
docker-compose up -d

# Проверка доступности PostgreSQL
echo "Ждём 10 секунд, пока PostgreSQL полностью запустится..."
sleep 10
docker exec m_postgres pg_isready
if [ $? -ne 0 ]; then
  echo "PostgreSQL не готов! Проверьте логи контейнера:"
  docker logs m_postgres
  error_exit
fi
echo "PostgreSQL запущен и готов к подключению."

# Установка зависимостей для backend
echo "Переходим в папку backend и устанавливаем зависимости Go..."
cd backend || { echo "Папка backend не найдена!"; error_exit; }

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
cd frontend || { echo "Папка frontend не найдена!"; error_exit; }

# Проверка, установлен ли npx (Create React App)
if ! command -v npx &> /dev/null
then
    echo "Ошибка: npx не найден. Устанавливаем Node.js, чтобы продолжить."
    npm install
fi

echo "Устанавливаем зависимые библиотеки..."
# Установка react-router-dom для маршрутизации и axios для HTTP-запросов.
npm install react-router-dom axios @mui/material @mui/icons-material @emotion/react @emotion/styled react-toastify
 || error_exit "Ошибка установки."


# Дополнительные библиотеки для валидации форм (при необходимости можно раскомментировать)
# npm install formik yup

# Возвращаемся в корневую директорию проекта
cd .. || error_exit "Не удалось вернуться в корень проекта."

echo "Окружение успешно настроено!"
echo "-------------------------------------------------"
echo "Для запуска приложения используйте команду: npm run dev"
echo "-------------------------------------------------"
