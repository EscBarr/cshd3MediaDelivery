FROM golang:1.26-alpine AS builder

WORKDIR /app

# Копируем весь код
COPY . .

# Если нет go.mod — создаём модуль и подтягиваем зависимости
RUN if [ ! -f go.mod ]; then \
        go mod init cshdMediaDelivery; \
    fi && \
    go mod tidy

# Сборка
RUN go build -o cshdMediaDelivery ./cmd/server/main.go


FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/cshdMediaDelivery /app/cshdMediaDelivery
COPY --from=builder /app/config-yaml /app/config-yaml

EXPOSE 6611

CMD ["./cshdMediaDelivery"]