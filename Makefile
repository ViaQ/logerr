include .bingo/Variables.mk

.DEFAULT_GOAL := all

all: lint test

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run ./...

.PHONY: test
test:
	go test -race ./...
