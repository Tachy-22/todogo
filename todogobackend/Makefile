.PHONY: build run setup-db

build:
	go mod tidy
	go build -o todogobackend

run: build
	./todogobackend

setup-db:
	psql -f schema.sql

dev:
	set -a && source .env && set +a && go run main.go