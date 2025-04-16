# Шаг сборки
FROM --platform=linux/amd64 golang:1.24.2 AS builder

# Устанавливаем полный набор зависимостей для libvips
RUN apt-get update && apt-get install -y \
    build-essential \
    pkg-config \
    libvips-dev \
    libmagickwand-dev \
    libmagickcore-dev \
    libvips-tools \
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

# Собираем бинарник (с поддержкой CGO для libvips)
# Явно указываем рабочую директорию и используем полный путь
WORKDIR /app/cmd
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 PKG_CONFIG_PATH=/usr/lib/pkgconfig go build -o /go/bin/app

# Шаг запуска - используем базовый образ с поддержкой glibc
FROM --platform=linux/amd64 debian:bookworm-slim
WORKDIR /root/

# Устанавливаем libvips и необходимые зависимости с ограничением памяти
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libvips42 \
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
