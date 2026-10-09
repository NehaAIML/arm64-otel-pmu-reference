.PHONY: all test clean ci build

all: test

test:
	go test -race ./internal/...

build:
	go build -o arm64-otel-pmu-collector ./cmd/agent

ci: test
	go vet ./...

clean:
	rm -f arm64-otel-pmu-reference.zip arm64-otel-pmu-collector
	go clean
