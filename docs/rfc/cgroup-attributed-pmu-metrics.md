# RFC: Cgroup-Attributed ARM64 PMU Metrics

## Summary
This RFC proposes a standardized method for exporting ARM64 Performance Monitoring Unit (PMU) counters attributed to Linux cgroups via OpenTelemetry.

## Motivation
As ARM64 (AWS Graviton, Ampere Altra) becomes dominant in cloud infrastructure, the lack of granular, low-overhead hardware visibility is a growing bottleneck. Existing solutions either lack container-level attribution or introduce prohibitive CPU overhead.

## Technical Approach
1. **eBPF Integration:** Use `PERF_EVENT_ARRAY` to read hardware counters directly in the kernel.
2. **Cgroup Attribution:** Utilize `bpf_get_current_cgroup_id()` to tag events at the source.
3. **Multiplexing Correction:** Implement userspace scaling (`enabled_delta / running_delta`) to ensure accuracy during event rotation.

## Semantic Conventions
We propose the following metric names:
- `arm64.pmu.cycles`: Total CPU cycles.
- `arm64.pmu.instructions`: Total instructions executed.
- `arm64.pmu.cache_misses`: Total L1/L2 cache misses.
