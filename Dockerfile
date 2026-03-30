# Используем официальный образ Go для сборки
FROM golang:1.22-alpine as builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git make

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o vpn-client ./cmd

# Используем минимальный базовый образ
FROM alpine:latest

# Устанавливаем ca-certificates для HTTPS
RUN apk --no-cache add ca-certificates

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранное приложение из builder
COPY --from=builder /app/vpn-client .

# Предоставляем портом
EXPOSE 8080 1080

# Устанавливаем переменную по умолчанию
ENV VPN_ADDR="0.0.0.0:8080"

# Запускаем приложение
CMD ["./vpn-client", "-addr", "0.0.0.0:8080"]
