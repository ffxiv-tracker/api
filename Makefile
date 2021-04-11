build:
	go build -o ./bin/api ./cmd/api

run:
	go run ./cmd/api

lint:
	staticcheck ffxiv.anid.dev/...

test:
	go test ffxiv.anid.dev/...

all: build