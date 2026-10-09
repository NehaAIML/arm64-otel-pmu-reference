# ARM64 PMU eBPF OpenTelemetry Collector

**Status:** Reference Implementation (Validation Pending on Target ARM64 Hardware)

> **The unsolved problem this project addresses**  
> On multi-tenant ARM64 nodes (AWS Graviton, Ampere Altra), operators still cannot answer:  
> **"Which container is causing cache thrashing or branch mispredictions at the silicon level?"**  
> Existing tools either stay at host aggregates, rely on energy models, or are x86-centric.  
> No open-source, low-overhead, cgroup-attributed, multiplexing-corrected stream of raw ARM64 PMU counters has been available — until now.

This repository provides a reference implementation: an in-kernel eBPF collector that reads real silicon PMU counters, attributes every sample to a Linux cgroup, corrects for hardware multiplexing in userspace, and exports standardized `arm64.pmu.*` metrics via OTLP.

---

## Why Existing Tools Do Not Solve This

| Tool | Granularity | Data Source | Missing Capability |
|------|-------------|-------------|--------------------|
| **Node Exporter** | Host only | `/proc`, `/sys` | No PMU counters; no cgroup attribution |
| **cAdvisor** | Container | cgroupfs | No access to PMU registers |
| **Kepler** | Pod (estimated) | Counters + Power Models | Exports derived energy estimates, not corrected raw PMU values |
| **Intel VTune** | Process | PMU (x86) | x86-only; proprietary; no OTLP |
| **AWS CloudWatch** | Instance | High-level metrics | No raw PMU register access |
| **This Project** | **Container (real)** | **Silicon PMU via eBPF** | **Cgroup attribution + multiplexing correction + OTLP** |

### Key Gaps Explained

- **Node Exporter** has zero visibility into hardware PMU events and cannot tag metrics with container cgroup IDs.
- **Kepler** consumes counters as *inputs* to energy models. Its "CPU Cycles" is a derived estimate, not ground-truth. This project's data can *improve* Kepler's models; they are complementary.
- **VTune / Arm Telemetry** are powerful but proprietary, x86-centric, or interactive. None ship an always-on, open-source, cgroup-attributed OTLP collector for multi-tenant Kubernetes.

---

## Architecture

```mermaid
graph TD
    subgraph Kernel
        A[eBPF Program] -->|Reads| B[PMU Registers]
        A -->|Tags via| C[bpf_get_current_cgroup_id]
        A -->|Stores in| D[PERCPU_HASH Map]
        E[Perf Event Arrays] -->|Feeds| A
    end
    subgraph Userspace
        F[Loader] -->|Manages| E
        F -->|Reads| D
        G[Collector] -->|Delta Calc| F
        G -->|Multiplexing Correction| H[Accumulator]
    end
    subgraph Backend
        I[OTLP Exporter] -->|gRPC| J[OTel Collector]
        H -->|Metrics| I
    end
