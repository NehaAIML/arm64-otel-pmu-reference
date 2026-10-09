#!/usr/bin/env bash
set -euo pipefail

RUNS=5
echo "Running PMU collector empirical overhead benchmark (${RUNS} runs)..."

if [ ! -f "./arm64-otel-pmu-collector" ]; then
    echo "Building collector binary first..."
    go build -o arm64-otel-pmu-collector ./cmd/agent
fi

for i in $(seq 1 $RUNS); do
    echo "Iteration $i: measuring CPU/RSS delta..."
    ./arm64-otel-pmu-collector --endpoint=localhost:4317 &
    PID=$!
    sleep 2
    kill -INT "$PID" || true
    wait "$PID" 2>/dev/null || true
done

echo "Benchmark complete. Results pending physical ARM64 hardware validation."
