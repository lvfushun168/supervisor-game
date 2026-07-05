.PHONY: dev-api dev-web build build-web tidy test mysql-up mysql-down

dev-api:
	go run .

dev-web:
	cd frontend && npm run dev

build: build-web
	go build -o bin/supervisor-game .

build-web:
	cd frontend && npm install && npm run build

tidy:
	go mod tidy

test:
	go test ./...

mysql-up:
	docker compose up -d mysql

mysql-down:
	docker compose down
