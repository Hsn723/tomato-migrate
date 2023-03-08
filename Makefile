PROJECT = tomato-migrate
VERSION = $(shell cat VERSION)
LDFLAGS=-ldflags "-w -s"

all: build

.PHONY: clean
clean:
	@if [ -f $(PROJECT) ]; then rm $(PROJECT); fi

generate:
	go install github.com/golang/mock/mockgen@latest
	go generate ./...

.PHONY: lint
lint: generate
	@if [ -z "$(shell which pre-commit)" ]; then pip3 install pre-commit; fi
	pre-commit install
	pre-commit run --all-files

.PHONY: test
test: generate
	go test --tags=test -coverprofile cover.out -count=1 -race -p 4 -v ./...

.PHONY: verify
verify:
	go mod download
	go mod verify

.PHONY: build
build: clean
	env CGO_ENABLED=0 go build $(LDFLAGS) .
