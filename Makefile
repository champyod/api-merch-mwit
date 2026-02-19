# Makefile for API-Merch-MWIT (MWIT-LINK Pattern)

DB_URL=postgres://user:password@localhost:5432/merch_db?sslmode=disable

.PHONY: dev build test migrate-diff migrate-apply docker-up docker-build docker-run docker-push

dev:
	go run main.go

build:
	go build -o bin/api main.go

test:
	go test ./...

migrate-diff:
	atlas migrate diff $(name) \
	  --dir "file://migrations" \
	  --to "gorm://localhost:5432/merch_db?sslmode=disable" \
	  --dev-url "docker://postgres/15/dev"

migrate-apply:
	atlas migrate apply \
	  --dir "file://migrations" \
	  --url "$(DB_URL)"

docker-up:
	docker compose up -d

docker-build:
	docker compose up -d --build api

docker-run:
	docker compose up api

docker-push:
	docker compose push api