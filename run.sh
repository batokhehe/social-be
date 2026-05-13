#!/bin/bash

echo "🧸 Running unit tests..."
go test ./...

if [ $? -ne 0 ]; then
  echo "❌ Tests failed. Docker build aborted."
  exit 1
fi

echo "🛑 Stopping containers..."
docker compose down

echo "🧹 Cleaning old build cache..."
docker builder prune -f

echo "🚀 Building & starting containers..."
docker compose up --build -d

echo "📋 Showing logs..."
docker compose logs -f