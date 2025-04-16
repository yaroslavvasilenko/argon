# Шаг сборки
FROM golang:1.24.1 AS builder

# Устанавливаем зависимости для libvips
RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    libvips-dev \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем рабочую директорию
WORKDIR /app

# Скопируем модули для сборки
COPY go.mod go.sum ./
RUN go mod download

# Скопируем весь проект
COPY . .

# Добавляем логирование для проверки структуры проекта
RUN echo "=== Содержимое корня проекта ===" && ls -la
RUN echo "=== Содержимое директории categories ===" && ls -la categories/

# Переходим в директорию, где лежит main.go
WORKDIR /app/cmd

# Собираем бинарник (с поддержкой CGO для libvips)
RUN CGO_ENABLED=1 GOOS=linux go build -o /go/bin/app

# Шаг запуска - используем базовый образ с поддержкой glibc
FROM debian:bookworm-slim
WORKDIR /root/

# Устанавливаем libvips и необходимые зависимости
RUN apt-get update && apt-get install -y \
    libvips-dev \
    libvips-tools \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Проверяем версию libvips
RUN vips --version

# Копируем бинарник из builder-контейнера
COPY --from=builder /go/bin/app .

# Копируем конфигурационные файлы
COPY --from=builder /app/config/config.toml ./config/

# Копируем директорию categories со всеми файлами
COPY --from=builder /app/categories ./categories/

# Копируем go.mod
COPY --from=builder /app/go.mod .

# Добавляем подробное логирование при сборке
RUN echo "Содержимое директории /root:" && ls -la
RUN echo "Содержимое директории /root/categories:" && ls -la ./categories/ || echo "Директория categories не существует"

# Проверяем версию libvips
RUN vips --version || echo "libvips не установлен или команда vips недоступна"

# Пробрасываем порт, на котором слушает Go-приложение
EXPOSE 8080

# Запускаем приложение
CMD ["./app"]
