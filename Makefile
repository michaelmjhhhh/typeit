.PHONY: build run test race vet fmt check generate grammars clean

build: grammars
	go build -trimpath -o bin/gittype ./cmd/gittype

run: grammars
	go run ./cmd/gittype

test: grammars
	go test ./...

race: grammars
	go test -race ./tests/go/...

vet: grammars
	go vet ./...

fmt:
	gofmt -w cmd internal assets/embed.go tests/go

check: grammars
	@test -z "$$(gofmt -l cmd internal assets/embed.go tests/go)"
	go vet ./...
	go test ./...
	go test -race ./tests/go/...

grammars:
	python3 scripts/port_grammars.py --ensure

generate:
	python3 scripts/port_grammars.py
	gofmt -w internal assets/embed.go

clean:
	rm -f bin/gittype
