BPF_SRC := bpf/pmu.bpf.c
BPF_OBJ := bpf/pmu.bpf.o
CMD := arm64-pmu-collector
IMAGE := quay.io/cilium/ebpf-builder:latest

.PHONY: all bpf build test vet fmt lint ci clean

all: bpf build

bpf: $(BPF_OBJ)

$(BPF_OBJ): $(BPF_SRC)
	@echo "Building eBPF object in Docker..."
	docker run --rm -v $(PWD):/src -w /src $(IMAGE) \
		clang -O2 -g -Wall -target bpf -c $(BPF_SRC) -o $(BPF_OBJ)

build: bpf
	go build -o $(CMD) ./cmd/arm64-pmu-collector

test:
	go test -race ./internal/...

vet:
	go vet ./...

fmt:
	gofmt -l -w ./cmd ./internal

lint: fmt vet
	@command -v golangci-lint >/dev/null && golangci-lint run ./... || echo "install golangci-lint"

ci: fmt vet test
	GOOS=linux GOARCH=arm64 go build -o /dev/null ./cmd/arm64-pmu-collector
	@echo "CI passed"

clean:
	rm -f $(BPF_OBJ) $(CMD)
