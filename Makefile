.PHONY: run test build docker

run:
	go run ./cmd/ghestimatebot/main.go

test:
	go test ./...

build:
	go build -o bin/ghestimatebot ./cmd/ghestimatebot/main.go

docker:
	docker build -t ghestimatebot:latest .
