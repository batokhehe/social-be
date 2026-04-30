# 🔹 Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# install git (dibutuhkan go run module)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 🔥 Generate swagger (FIX)
RUN go run github.com/swaggo/swag/cmd/swag@v1.16.2 init -g cmd/main.go

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o app ./cmd

# 🔹 Run stage
FROM alpine:3.22

WORKDIR /root/

RUN adduser -D -H appuser && apk add --no-cache ca-certificates

# copy binary + swagger docs
COPY --from=builder /app/app .
COPY --from=builder /app/docs ./docs

USER appuser

EXPOSE 8080

CMD ["./app"]