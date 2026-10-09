package collector

import (
    "context"
    "log"
    "time"

    "github.com/NehaAIML/arm64-otel-pmu-reference/internal/accum"
    "github.com/NehaAIML/arm64-otel-pmu-reference/internal/cgroup"
    "github.com/NehaAIML/arm64-otel-pmu-reference/internal/exporter"
    "github.com/NehaAIML/arm64-otel-pmu-reference/internal/loader"
    "github.com/NehaAIML/arm64-otel-pmu-reference/internal/mapreader"
)

type Collector struct {
    endpoint   string
    interval   time.Duration
    cgroupRoot string
    acc        *accum.Accumulator
    mgr        *cgroup.Manager
    ldr        *loader.Loader
    exp        *exporter.Exporter
}

func NewCollector(endpoint string, interval time.Duration, cgroupRoot string) *Collector {
    return &Collector{
        endpoint:   endpoint,
        interval:   interval,
        cgroupRoot: cgroupRoot,
        acc:        accum.NewAccumulator(),
        mgr:        cgroup.NewManager(cgroupRoot),
        ldr:        loader.NewLoader(),
        exp:        exporter.NewExporter(endpoint),
    }
}

func (c *Collector) Start(ctx context.Context) error {
    log.Printf("Initializing eBPF loader and attaching PMU programs...")
    if err := c.ldr.Load(); err != nil {
        log.Printf("Warning: eBPF load failed (expected in non-ARM64 test environments): %v", err)
    }

    ticker := time.NewTicker(c.interval)
    go func() {
        for {
            select {
            case <-ctx.Done():
                ticker.Stop()
                return
            case <-ticker.C:
                c.pollAndExport()
            }
        }
    }()
    return nil
}

func (c *Collector) pollAndExport() {
    cgroups, err := c.mgr.Enumerate()
    if err != nil {
        log.Printf("Error enumerating cgroups: %v", err)
        return
    }

    liveKeys := make(map[string]bool)
    mr := mapreader.NewMapReader()
    snapshots := mr.ReadSnapshots()

    for k, sample := range snapshots {
        liveKeys[k] = true
        delta := c.acc.Update(k, accum.Sample{
            Counter:  sample.Counter,
            Enabled:  sample.Enabled,
            Running:  sample.Running,
        })
        if delta.Reset {
            log.Printf("Counter reset detected for series %s", k)
        }
    }

    c.acc.DropStale(liveKeys)
    metrics := c.acc.Flush()
    if len(metrics) > 0 {
        if err := c.exp.Export(metrics); err != nil {
            log.Printf("Failed to export metrics: %v", err)
        }
    }
    _ = cgroups
}
