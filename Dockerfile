FROM golang:latest as dev

WORKDIR /app

CMD ["go run cmd/main/main.go"]
