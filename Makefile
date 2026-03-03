.PHONY: build run-memory run-postgres test generate docker-build docker-up docker-up-memory docker-down docker-logs migrate-up migrate-down

BIN        := ./bin/url-shortener
DOCKER_TAG := url-shortener:latest

build:
	go build -o $(BIN) ./cmd/server

run-memory: build
	$(BIN)

run-postgres: build
	$(BIN) -config=configs/postgres.yaml

test:
	go test -race -count=1 ./...

generate:
	sqlc generate
	go generate ./...

docker-build:
	docker build -t $(DOCKER_TAG) .

docker-up:
	docker compose up --build -d postgres migrate app-postgres

docker-up-memory:
	docker compose --profile memory up --build -d app-memory

migrate-up:
	docker compose run --rm migrate -path=/migrations -database "postgres://shortener:shortener@postgres:5432/shortener?sslmode=disable" up

migrate-down:
	docker compose run --rm migrate -path=/migrations -database "postgres://shortener:shortener@postgres:5432/shortener?sslmode=disable" down 1

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app-postgres
