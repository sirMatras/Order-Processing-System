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
RUN go build -o order-service main.go

# Открываем порт для приложения
EXPOSE 8083

# Запускаем приложение
CMD ["./order-service"]
