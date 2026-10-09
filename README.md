# ARM64 PMU eBPF OpenTelemetry Collector

**Status**: Reference Implementation (Validation Pending on Target ARM64 Hardware)

A low-level Performance Monitoring Unit (PMU) eBPF metric collector targeting ARM64 Linux kernels, exporting standardized `arm64.pmu.*` metrics via OTLP gRPC.

## Architecture
- **eBPF Kernel Component (`bpf/pmu.bpf.c`)**: Dual BSD/GPL licensed (`SPDX-License-Identifier: (GPL-2.0-only OR BSD-2-Clause)`). Collects cycle counts, instructions, and cache misses using 9-field per-CPU snapshot maps and dedicated perf event arrays.
- **Collector Engine (`internal/collector`)**: Polls per-CPU snapshot maps via `cilium/ebpf`, applies 10x `MaxScale` multiplexing correction, and manages monotonic delta accumulation with reset detection.
- **Export Pipeline (`internal/exporter`)**: OTLP gRPC metric transmission conforming to OpenTelemetry semantic conventions.

## Validation Status
*Note: Per ADR-005, performance and overhead claims are withheld pending empirical benchmarking on physical ARM64 hardware (AWS Graviton3 / Qualcomm Ampere).*

## Quick Start & Benchmarking
```bash
make test
./bench-overhead.sh

mkdir -p "$WORKSPACE/bpf" "$WORKSPACE/cmd/agent" "$WORKSPACE/internal/perf"

cat << 'INNER' > "$WORKSPACE/bpf/pmu.bpf.c"
// SPDX-License-Identifier: (GPL-2.0-only OR BSD-2-Clause)
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

char __license[] SEC("license") = "Dual BSD/GPL";

struct pmu_sample {
    __u64 cycles_counter;
    __u64 cycles_enabled;
    __u64 cycles_running;
    __u64 instructions_counter;
    __u64 instructions_enabled;
    __u64 instructions_running;
    __u64 misses_counter;
    __u64 misses_enabled;
    __u64 misses_running;
};

struct map_key {
    __u32 cpu;
    __u64 cgroup_id;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_HASH);
    __uint(key_size, sizeof(struct map_key));
    __uint(value_size, sizeof(struct pmu_sample));
    __uint(max_entries, 1024);
} pmu_snapshots SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(int));
    __uint(max_entries, 256);
} cycles_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(int));
    __uint(max_entries, 256);
} instructions_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(int));
    __uint(max_entries, 256);
} cache_misses_map SEC(".maps");

SEC("perf_event")
int handle_pmu_event(struct bpf_perf_event_data *ctx) {
    __u32 cpu = bpf_get_smp_processor_id();
    struct map_key key = {
        .cpu = cpu,
        .cgroup_id = bpf_get_current_cgroup_id(),
    };

    struct pmu_sample sample = {};
    bpf_perf_event_read_value(&cycles_map, cpu, &sample.cycles_counter, sizeof(__u64) * 3);
    bpf_perf_event_read_value(&instructions_map, cpu, &sample.instructions_counter, sizeof(__u64) * 3);
    bpf_perf_event_read_value(&cache_misses_map, cpu, &sample.misses_counter, sizeof(__u64) * 3);

    bpf_map_update_elem(&pmu_snapshots, &key, &sample, BPF_ANY);
    return 0;
}
