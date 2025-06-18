# Используем официальный образ Golang для сборки
FROM golang:1.24-alpine

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем исполняемый файл
RUN go build -o payment-service main.go

# Добавляем healthcheck для сервиса
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD curl --fail http://localhost:8082/health || exit 1

# Открываем порт для приложения
EXPOSE 8082

# Запускаем приложение
CMD ["./payment-service"]
