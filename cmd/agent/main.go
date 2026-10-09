package main

import (
    "context"
    "flag"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/NehaAIML/arm64-otel-pmu-reference/internal/collector"
)

func main() {
    endpoint := flag.String("endpoint", "localhost:4317", "OTLP gRPC collector endpoint")
    interval := flag.Duration("interval", 10*time.Second, "Polling interval")
    cgroupRoot := flag.String("cgroup-root", "/sys/fs/cgroup", "Cgroup filesystem root")
    flag.Parse()

    log.Println("Starting ARM64 PMU eBPF OpenTelemetry Collector...")
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    c := collector.NewCollector(*endpoint, *interval, *cgroupRoot)
    if err := c.Start(ctx); err != nil {
        log.Fatalf("Collector failed to start: %v", err)
    }

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    log.Println("Shutting down collector gracefully...")
    cancel()
}
