.PHONY: migrate sqlc test run

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/ocealis?sslmode=disable

migrate:
	psql "$(DATABASE_URL)" -f db/schema.sql

sqlc:
	cd db && sqlc generate

test:
	go test ./...

run:
	go run .
