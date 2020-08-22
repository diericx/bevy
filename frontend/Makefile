.ONESHELL:
SHELL   := /bin/bash

PKGS 	:= $(shell go list ./... | grep -v /vendor/)
SOURCE_DIRS := internal cmd test tools
VERSION ?= $(shell git rev-list --count HEAD).$(shell git rev-parse --short HEAD)

.DEFAULT_GOAL = all

.SUFFIXES:

.PHONY: all
all: deps build

.PHONY: docker
docker:
	@docker build -t iceetime/frontend:$(VERSION) -f Dockerfile .

.PHONY: clean
clean:
	@rm -fr dist/*
