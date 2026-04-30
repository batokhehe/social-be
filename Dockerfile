# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o app ./cmd

# Run stage
FROM alpine:3.22

WORKDIR /root/

RUN adduser -D -H appuser && apk add --no-cache ca-certificates

COPY --from=builder /app/app .

USER appuser

EXPOSE 8080

CMD ["./app"]
