FROM golang:latest as dev

WORKDIR /app

# Копируем go.mod и go.sum (если они есть)
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код приложения
COPY . .

# Указываем команду для запуска
CMD ["go", "run", "cmd/main/main.go"]