
.PHONY: build lint test

build:
	@go build

cleanbuild:
	@go clean
	@go build

lint:
	@golangci-lint run

test:
	@go test ./...