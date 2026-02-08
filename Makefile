.PHONY: build test lint fmt deps gotestsum test-update

build:
	go build -race -o bin/terraform-gitlab-drift .

test:
	go test -v -race -count=1 ./...

lint:
	golangci-lint run -v

fmt:
	go fmt ./...

deps:
	go mod download
	go mod tidy
	go mod verify

gotestsum:
	gotestsum --watch -- --count=1 --timeout=5s

test-update:
	UPDATE_GOLDEN=1 go test -v -count=1 ./...
