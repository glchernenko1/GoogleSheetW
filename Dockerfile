# Используем официальный образ Go
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY app/go.mod app/go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY app/ .

# Собираем приложение
RUN GOOS=linux go build -o main ./cmd/main.go

# Используем минимальный образ для production
FROM alpine:latest

WORKDIR /app

# Копируем собранное приложение
COPY --from=builder /app/main .

# Создаем директорию для логов
RUN mkdir -p /app/log

# Запускаем приложение
CMD ["./main"]
