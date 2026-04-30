#!/bin/bash

echo "🛑 Stopping containers..."
docker compose down

echo "🧹 Cleaning old build cache..."
docker builder prune -f

echo "🚀 Building & starting containers..."
docker compose up --build -d

echo "📋 Showing logs..."
docker compose logs -f