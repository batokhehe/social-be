#!/bin/sh

DB_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

echo "Running migration..."
echo "DBURL: $DB_URL"
migrate -path migrations -database "$DB_URL" up

echo "Migration done"