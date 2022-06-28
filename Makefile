.PHONY: all install mod clean vet test

# Target all gofiles
TARGETS = ./...

all: mod clean vet test install

install:
	go install $(TARGETS)

mod:
	go mod tidy

clean:
	go clean 

vet:
	go vet $(TARGETS)
	golangci-lint run --enable-all

test: 
	go test -v $(TARGETS)
