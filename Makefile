all: run

run:
	go run ./cmd/

build:
	go build -o app ./cmd/
