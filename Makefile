# Makefile for API-Merch-MWIT (MWIT-LINK Pattern)

DB_URL=postgres://user:password@localhost:5432/merch_db?sslmode=disable

.PHONY: dev build test migrate-diff migrate-apply docker-build docker-run

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

docker-build:
	docker build -t api-merch-mwit .

docker-run:
	docker run -p 8080:8080 --env-file .env.local api-merch-mwit

docker-push:
	docker push api-merch-mwit