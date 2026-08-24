BINARY := nx-cache-service
BIN_DIR := bin
PKG := ./cmd/nx-cache-service
SPEC := api/openapi.yaml
GO ?= go

.DEFAULT_GOAL := build
.PHONY: all build run test test-race cover fmt fmt-check vet tidy spec-lint check clean

all: check build

build:
	$(GO) build -o $(BIN_DIR)/$(BINARY) $(PKG)

run:
	$(GO) run $(PKG)

test:
	$(GO) test ./...

test-race:
	$(GO) test -count=1 -race ./...

cover:
	$(GO) test -count=1 -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

fmt:
	gofmt -w .

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "not gofmt-ed:"; echo "$$unformatted"; exit 1; \
	fi

vet:
	$(GO) vet ./...

spec-lint:
	npx --yes @redocly/cli@latest lint $(SPEC)

tidy:
	$(GO) mod tidy

check: fmt-check vet test-race

clean:
	rm -rf $(BIN_DIR) coverage.out
