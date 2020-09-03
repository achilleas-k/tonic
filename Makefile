SOURCES = $(shell find . -type f -iname "*.go") go.mod go.sum

.PHONY: clean test coverreport

all: utonics

utonics: $(SOURCES)
	mkdir -p build
	go build -v -o ./build ./utonics/...

test: $(SOURCES)
	go test -race -coverpkg=./... -coverprofile=coverage ./...

coverreport: test
	go tool cover -html=coverage -o coverage.html

clean:
	rm -rf coverage build/
