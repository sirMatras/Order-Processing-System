# Используем официальный образ Golang для сборки
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем исполняемый файл
RUN go build -o /api-gateway main.go

# Используем минимальный образ для запуска
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /root/

# Копируем собранный бинарный файл из этапа сборки
COPY --from=builder /api-gateway .

# Открываем порт для API
EXPOSE 8080

# Запускаем приложение
CMD ["./api-gateway"]
