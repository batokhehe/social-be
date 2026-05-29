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

echo "📌 Mounting NAS inside app container..."
docker compose exec -T app ./mount-nas.sh

if [ $? -ne 0 ]; then
  echo "❌ NAS mount failed. Stopping containers..."
  docker compose down
  exit 1
fi

echo "🔎 Verifying NAS mount..."
docker compose exec -T app sh -c 'mountpoint -q "${NAS_MOUNT_PATH:-/mnt/nas}"'

if [ $? -ne 0 ]; then
  echo "❌ NAS mount verification failed. Stopping containers..."
  docker compose down
  exit 1
fi

echo "✓ NAS mounted successfully"

echo "📋 Showing logs..."
docker compose logs -f
