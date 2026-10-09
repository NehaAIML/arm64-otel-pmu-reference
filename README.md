# ARM64 PMU eBPF OpenTelemetry Collector

**Status:** Reference Implementation (Validation Pending on Target ARM64 Hardware)

A low-level Performance Monitoring Unit (PMU) eBPF metric collector targeting ARM64 Linux kernels, exporting standardized `arm64.pmu.*` metrics via OTLP gRPC.

## Architecture

*   **eBPF Kernel Component (`bpf/pmu.bpf.c`):** Dual BSD/GPL licensed. Collects cycle counts, instructions, and cache misses using per-CPU snapshot maps and dedicated `PERF_EVENT_ARRAY`s.
*   **Collector Engine (`internal/collector`):** Polls per-CPU snapshot maps via `cilium/ebpf`, applies multiplexing correction (`enabled_delta / running_delta`), and manages monotonic delta accumulation with reset detection.
*   **Export Pipeline (`internal/otlp`):** OTLP gRPC metric transmission conforming to OpenTelemetry semantic conventions.

## Validation Status

> **Note:** Per ADR-005, performance and overhead claims are withheld pending empirical benchmarking on physical ARM64 hardware (AWS Graviton3 / Ampere Altra). Current CI validation uses software events on x86_64 runners to verify build integrity and logic correctness.

## Quick Start

### Prerequisites
*   Linux Kernel ≥ 5.15 (for stable eBPF support)
*   Clang ≥ 12
*   Go ≥ 1.22

### Build & Run
```bash
# 1. Compile the eBPF object
make bpf

# 2. Build the Go collector
make build

# 3. Run the collector (requires root/CAP_SYS_ADMIN)
sudo ./arm64-pmu-collector --endpoint localhost:4317
