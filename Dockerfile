# 🔹 Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Run repository-wide unit tests before building the image
RUN go test ./...

# Generate swagger
RUN go run github.com/swaggo/swag/cmd/swag@v1.16.2 init -g cmd/main.go

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o app ./cmd

# 🔹 Run stage
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates bash curl

# install migrate CLI
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.3/migrate.linux-amd64.tar.gz \
    | tar -xz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# copy files
COPY --from=builder /app/app .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/migrate.sh .

# permission
RUN chmod +x ./app ./migrate.sh

EXPOSE 8080

CMD ["sh", "-c", "./migrate.sh && ./app"]