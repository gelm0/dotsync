.PHONY: all install build run mod clean vet test

# Target all gofiles
TARGETS = ./...

all: mod clean vet test build

install:
	go install $(TARGETS)

build:
	go build $(TARGETS)

run:
	go run $(TARGETS)

mod:
	go mod tidy

clean:
	go clean 

vet:
	go vet $(TARGETS)
	golangci-lint run --enable-all

test: 
	go test -v $(TARGETS)
