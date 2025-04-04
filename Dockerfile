# Шаг сборки
FROM golang:1.24.1 AS builder

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

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/app

# Шаг запуска
FROM alpine:latest
WORKDIR /root/

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

# Пробрасываем порт, на котором слушает Go-приложение
EXPOSE 8080

# Запускаем приложение
CMD ["./app"]
