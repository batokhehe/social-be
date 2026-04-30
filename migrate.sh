#!/bin/bash

export $(grep -v '^#' .env | xargs)

DB_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

migrate -path migrations -database "$DB_URL" up