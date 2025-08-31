.PHONY: run test build docker

run:
	go run ./cmd/ghbot/main.go

test:
	go test ./...

build:
	go build -o bin/ghbot ./cmd/ghbot/main.go

docker:
	docker build -t ghbot:latest .
