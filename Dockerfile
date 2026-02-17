# Первый этап: сборка бинаря
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копируем модули и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Копируем весь исходный код
COPY . .

# Сборка только cshdMediaDelivery
RUN go build -o cshdMediaDelivery ./cmd/server/main.go

# Второй этап: минимальный образ для запуска
FROM alpine:latest

WORKDIR /app

# Добавляем сертификаты для https-запросов
RUN apk add --no-cache ca-certificates

# Копируем только бинарь и конфиг
COPY --from=builder /app/cshdMediaDelivery /app/cshdMediaDelivery
COPY --from=builder /app/config-yaml /app/config-yaml

# Открываем порт (замени на тот, который слушает твой Gateway)
EXPOSE 6611

# # Запускаем бинарь
# CMD ["./api-gateway"]
