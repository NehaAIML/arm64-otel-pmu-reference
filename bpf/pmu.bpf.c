// SPDX-License-Identifier: GPL-2.0-only OR BSD-2-Clause
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
char LICENSE[] SEC("license") = "Dual BSD/GPL";

struct { __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY); __uint(max_entries, 1024); __type(key, __u32); __type(value, __u32); } cycles_events SEC(".maps");
struct { __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY); __uint(max_entries, 1024); __type(key, __u32); __type(value, __u32); } instr_events SEC(".maps");
struct { __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY); __uint(max_entries, 1024); __type(key, __u32); __type(value, __u32); } miss_events SEC(".maps");

struct pmu_key { __u32 cpu; __u64 cgroup_id; };
struct pmu_val {
	__u64 cycles; __u64 cycles_en; __u64 cycles_run;
	__u64 instr; __u64 instr_en; __u64 instr_run;
	__u64 miss; __u64 miss_en; __u64 miss_run;
};
struct { __uint(type, BPF_MAP_TYPE_PERCPU_HASH); __uint(max_entries, 16384); __type(key, struct pmu_key); __type(value, struct pmu_val); } pmu_snapshots SEC(".maps");

SEC("perf_event") int collect_pmu_counters(void *ctx) {
	__u32 cpu = bpf_get_smp_processor_id();
	struct bpf_perf_event_value pev = {};
	struct pmu_key key = {.cpu=cpu, .cgroup_id=bpf_get_current_cgroup_id()};
	struct pmu_val *v, zero={}; long ret;
	v = bpf_map_lookup_elem(&pmu_snapshots, &key);
	if (!v) { bpf_map_update_elem(&pmu_snapshots,&key,&zero,BPF_NOEXIST); v=bpf_map_lookup_elem(&pmu_snapshots,&key); if(!v) return 0; }
	ret=bpf_perf_event_read_value(&cycles_events,cpu,&pev,sizeof(pev)); if(ret==0){v->cycles=pev.counter;v->cycles_en=pev.enabled;v->cycles_run=pev.running;}
	ret=bpf_perf_event_read_value(&instr_events,cpu,&pev,sizeof(pev)); if(ret==0){v->instr=pev.counter;v->instr_en=pev.enabled;v->instr_run=pev.running;}
	ret=bpf_perf_event_read_value(&miss_events,cpu,&pev,sizeof(pev)); if(ret==0){v->miss=pev.counter;v->miss_en=pev.enabled;v->miss_run=pev.running;}
	return 0;
}
