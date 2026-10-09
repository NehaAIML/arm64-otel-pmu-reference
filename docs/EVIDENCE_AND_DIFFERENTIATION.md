# Evidence & Differentiation: Why This is Unique

## Claim 1: "Prometheus Node Exporter already does this."
**Evidence:** Node Exporter reads `/proc` and `/sys`. It provides **host-level** aggregates. It cannot attribute hardware events to specific Kubernetes pods. 
**Proof:** Run `node_exporter` on a multi-tenant Graviton node; you will see total CPU usage but no insight into which container is causing L1 cache misses.

## Claim 2: "Kepler already tracks performance."
**Evidence:** Kepler is an **energy estimator**. It uses performance counters as *inputs* to a power model. It does not export the raw, accurate hardware counters themselves.
**Proof:** Compare Kepler's "CPU Cycles" metric against our raw eBPF output. Ours is the ground truth; Kepler's is a derived estimate optimized for carbon tracking, not silicon debugging.

## Claim 3: "Intel and AWS have their own tools."
**Evidence:** Intel's tools (VTune) are heavy, proprietary, and x86-centric. AWS CloudWatch provides high-level metrics but lacks the granularity of raw PMU registers.
**Proof:** Neither provides an open-source, OTLP-compliant, cgroup-attributed stream of raw PMU data for ARM64.

## Who is this useful for?
1. **Platform Engineers:** Identifying "noisy neighbors" in shared ARM64 clusters.
2. **Performance Teams:** Debugging why a specific microservice is slow on Graviton vs. x86.
3. **FinOps Teams:** Correlating actual silicon resource usage with cloud billing data.
