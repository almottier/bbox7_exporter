# Go 1.21 produces a broken macOS binary (missing LC_UUID): force a recent Go.
GOTOOLCHAIN ?= go1.23.4
export GOTOOLCHAIN

BINARY  := bbox7_exporter
ENDPOINT ?= https://mabbox.bytel.fr

.PHONY: build run fmt vet tidy clean

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY) --endpoint=$(ENDPOINT)

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
