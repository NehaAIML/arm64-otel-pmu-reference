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
