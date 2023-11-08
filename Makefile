.PHONY: build test

build:
	@go build -o build/devserver ./cmd

run:
	@go run ./cmd

test:
	@go test -v ./test/...