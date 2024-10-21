SHELL := bash

.PHONY: all
all: lint test coverage check-licenses build

.PHONY: build
build: .

.PHONY: lint
lint: prepare
	golangci-lint run

.PHONY: prepare
prepare:
	go mod vendor
	
.PHONY: test
test: prepare
	go test ./...

.PHONY: coverage
coverage:
	# Ignore (allow) packages without any tests
	set -o pipefail
	go test ./... -coverprofile coverage.out
	set +o pipefail
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func coverage.out -o coverage.txt
	tail -1 coverage.txt



.PHONY: clean
clean:
	rm -f out/*
	rm -r examples-out/*

.PHONY: install
install: all
	cp out/* $(GOPATH)/bin/

.PHONY: update
update:
	go get -t -u ./...

.PHONY: check-licenses
check-licenses: prepare
	wwhrd check
