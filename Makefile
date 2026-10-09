BPF_SRC := bpf/pmu.bpf.c
BPF_OBJ := bpf/pmu.bpf.o
CMD := arm64-pmu-collector

.PHONY: all bpf build test vet fmt ci clean

all: bpf build

bpf: $(BPF_OBJ)

$(BPF_OBJ): $(BPF_SRC)
	clang -O2 -g -Wall -target bpf -c $< -o $@

build: bpf
	go build -o $(CMD) ./cmd/arm64-pmu-collector

test:
	go test -race ./internal/...

vet:
	go vet ./...

fmt:
	gofmt -l -w ./cmd ./internal

ci: fmt vet test
	GOOS=linux GOARCH=arm64 go build -o /dev/null ./cmd/arm64-pmu-collector
	@echo "CI passed"

clean:
	rm -f $(BPF_OBJ) $(CMD)
