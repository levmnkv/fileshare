# Этап сборки (Builder stage)
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем go.mod
COPY go.mod ./

# Если появятся зависимости, раскомментируйте:
# COPY go.sum ./
# RUN go mod download

# Копируем исходный код
COPY main.go ./

# Собираем бинарник под Raspberry Pi 4 (ARM64)
# CGO_ENABLED=0 обеспечивает статическую линковку
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o fileshare main.go

# Финальный этап (Final stage)
FROM alpine:latest

WORKDIR /app

# Создаем необходимые директории
RUN mkdir -p uploads static

# Копируем собранный бинарник из этапа сборки
COPY --from=builder /app/fileshare .

# Копируем статические файлы фронтенда
COPY static/ ./static/

# Открываем порт
EXPOSE 8080

# Запускаем приложение
CMD ["./fileshare"]
