BINARY := nx-cache-service
BIN_DIR := bin
PKG := ./cmd/nx-cache-service
GO ?= go

.DEFAULT_GOAL := build
.PHONY: all build run fmt fmt-check vet tidy check clean

all: check build

build:
	$(GO) build -o $(BIN_DIR)/$(BINARY) $(PKG)

run:
	$(GO) run $(PKG)

fmt:
	gofmt -w .

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "not gofmt-ed:"; echo "$$unformatted"; exit 1; \
	fi

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

check: fmt-check vet

clean:
	rm -rf $(BIN_DIR)
