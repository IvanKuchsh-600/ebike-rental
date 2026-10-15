# ============ Build stage ============
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod/go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/seed ./cmd/seed

# ============ Runtime stage ============
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Копируем бинарники
COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/seed /app/seed

# Копируем шаблоны и статику
COPY --from=builder /app/web /app/web

EXPOSE 8080

CMD ["/app/api"]