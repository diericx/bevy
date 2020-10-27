.ONESHELL:
SHELL   := /bin/bash

PKGS 	:= $(shell go list ./... | grep -v /vendor/)
SOURCE_DIRS := internal cmd test tools
VERSION ?= $(shell git rev-list --count HEAD).$(shell git rev-parse --short HEAD)

.DEFAULT_GOAL = all

.SUFFIXES:

.PHONY: all
all: deps build

.PHONY: build
build: macos linux docker

.PHONY: docker
docker:
	@docker build -t iceetime/iceetime:$(VERSION) .

.PHONY: linux-amd64
linux-amd64:
	@mkdir -p dist/linux
	@cd dist/linux; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-s -w -extldflags "-static" -X=main.version=$(VERSION)' -o linux-amd64 ../../cmd/...

.PHONY: linux-arm64
linux-arm64:
	@mkdir -p dist/linux
	@cd dist/linux; CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -ldflags '-s -w -extldflags "-static" -X=main.version=$(VERSION)' -o linux-arm64 ../../cmd/...

.PHONY: macos
macos:
	@mkdir -p dist/macos
	@cd dist/macos; go build -ldflags '-X=main.version=$(VERSION)' ../../cmd/...

.PHONY: deps
deps:
	@go mod download

.PHONY: test
test:
	@go test -v $(PKGS)

.PHONY: bench
bench:
	@go test -bench=. -v $(PKGS)

.PHONY: lint
lint:
	@go vet -v $(PKGS)

.PHONY: checkfmt
checkfmt:
	@gofmt -l $(SOURCE_DIRS) | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi

.PHONY: fmt
fmt:
	@gofmt -l -w $(SOURCE_DIRS)

.PHONY: check
check: lint test

.PHONY: clean
clean:
	@rm -fr dist/*

.PHONY: locust-client
locust-client:
	@mkdir -p test/locust/dist
	@cd test/locust/dist; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-s -w -extldflags "-static" -X=main.version=$(VERSION)' ../client/...

mocks:
	mockgen -source=internal/app/app.go -destination internal/app/mocks/app_mocks.go -package mocks && \
	mockgen -source=internal/pkg/torrent/client.go -destination internal/app/mocks/torrents_mocks.go -package mocks
