.PHONY: build test run

build:
	go build -o bin/seeding-bot ./cmd/seeding-bot

test:
	go test ./...

run: build
	./bin/seeding-bot
