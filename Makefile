PKG_LIST := $(shell go list ./...)
GO_FILES := $(shell find . -name '*.go' | grep -v _test.go)

.PHONY: build
build:
	go build -race -o bin/terraform-gitlab-drift main.go

.PHONY: run
run: build
	./bin/terraform-gitlab-drift

.PHONY: lint-test
lint-test:
	golangci-lint run -v

.PHONY: staticcheck-test
staticcheck-test:
	staticcheck ${PKG_LIST}

.PHONY: fmt-test
fmt-test:
	go fmt ${PKG_LIST}

.PHONY: vet-test
vet-test:
	go vet ${PKG_LIST}

.PHONY: unit-test
unit-test:
	go test -v ${PKG_LIST} -count=1 -timeout=10s

.PHONY: race-test
race-test:
	go test -race -short ${PKG_LIST}  -count=1 -timeout=10s

.PHONY: gosec-test
gosec-test:
	gosec ${PKG_LIST}

.PHONY: benchmark-test
benchmark-test:
	go test -bench=. -benchmem ${PKG_LIST}

.PHONY: gotestsum
gotestsum:
	gotestsum --watch -- --count=1 --timeout=5s
