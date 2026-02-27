# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем минимальный бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Production stage (легковесный образ)
FROM alpine:latest

WORKDIR /app

# Копируем готовый бинарник из стадии сборки
COPY --from=builder /app/main .

EXPOSE 8080

# Запускаем приложение
CMD ["./main"]
