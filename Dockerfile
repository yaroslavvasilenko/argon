# Шаг сборки
FROM golang:1.23.2 AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Скопируем модули для сборки
COPY go.mod go.sum ./
RUN go mod download

# Скопируем весь проект
COPY . .

# Переходим в директорию, где лежит main.go
WORKDIR /app/cmd

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/app

# Шаг запуска
FROM alpine:latest
WORKDIR /root/



COPY --from=builder /go/bin/app .
COPY --from=builder /app/config/config.toml ./config/
COPY --from=builder /app/categories.json .

# Пробрасываем порт, на котором слушает Go-приложение
EXPOSE 8080

# Запускаем приложение
CMD ["./app"]
