 Установка Docker в WSL2 (Ubuntu) через терминал

📦 Шаг 1: Обновляем пакеты
bash
Копировать
Редактировать
sudo apt update && sudo apt upgrade -y
🐳 Шаг 2: Установка зависимостей Docker
bash
Копировать
Редактировать
sudo apt install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release -y
🔐 Шаг 3: Добавление GPG-ключа Docker
bash
Копировать
Редактировать
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | \
  sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
📁 Шаг 4: Добавление Docker-репозитория
bash
Копировать
Редактировать
echo \
  "deb [arch=$(dpkg --print-architecture) \
  signed-by=/etc/apt/keyrings/docker.gpg] \
  https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
🔄 Шаг 5: Установка Docker Engine
bash
Копировать
Редактировать
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y
✅ Шаг 6: Проверка установки
bash
Копировать
Редактировать
docker --version
👤 Шаг 7: (Опционально) Разрешить запуск Docker без sudo
bash
Копировать
Редактировать
sudo usermod -aG docker $USER
После этого перезапусти терминал или введи:

bash
Копировать
Редактировать
newgrp docker
🐧 Как запустить Docker в WSL2:
WSL не запускает Docker как демон по умолчанию. Вместо этого:

✅ Рекомендуемый способ: Использовать Docker Desktop на Windows
Установи Docker Desktop: https://www.docker.com/products/docker-desktop/

Включи «WSL integration» в настройках Docker Desktop.

Запусти Docker Desktop на Windows, он поднимет Docker для WSL.

→ Теперь в WSL2 можно использовать Docker как обычно (docker run, docker compose, и т.д.)

🔍 Проверка
bash
Копировать
Редактировать
docker run hello-world 