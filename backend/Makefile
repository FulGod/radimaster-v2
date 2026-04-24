.PHONY: dev build clean

dev:
	go run cmd/server/main.go

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/radimaster cmd/server/main.go

clean:
	rm -rf bin/ radimaster.db
