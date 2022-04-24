.PHONY: all test clean

clean:
	go clean
	go mod tidy

test:
	go test -v

install:
	go test
	go vet
	go install
