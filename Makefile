.PHONY: dev-db build test lint clean up down

dev-db:
	docker compose -f docker-compose.dev.yml up -d

build:
	cd backend && go build -trimpath -o bin/server ./cmd/server
	cd frontend && npm run build

test:
	cd backend && go test ./...
	cd frontend && npm run lint

lint:
	cd backend && golangci-lint run ./...
	cd frontend && npm run lint && npm run format:check

clean:
	cd backend && rm -rf bin/
	cd frontend && rm -rf dist/ node_modules/

up:
	docker compose up --build -d

down:
	docker compose down
