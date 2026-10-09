//go:build linux

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NehaAIML/arm64-otel-pmu-reference/internal/bpf"
	"github.com/NehaAIML/arm64-otel-pmu-reference/internal/collector"
	"github.com/NehaAIML/arm64-otel-pmu-reference/internal/otlp"
	"github.com/cilium/ebpf"
)

func main() {
	poll := flag.Duration("poll", 1*time.Second, "Poll interval")
	export := flag.Duration("export", 10*time.Second, "Export interval")
	endpoint := flag.String("endpoint", "localhost:4317", "OTLP endpoint")
	bpfObj := flag.String("bpf", "bpf/pmu.bpf.o", "eBPF object path")
	flag.Parse()

	pc, err := bpf.NewPMUCollector(*bpfObj)
	if err != nil {
		log.Fatalf("loader: %v", err)
	}
	defer pc.Close()

	ncpu := bpf.GetNumCPUs()
	reader := bpf.NewMapReader(pc.Snapshots(), ncpu)
	exp := otlp.NewExporter(*endpoint, false)

	cfg := collector.Config{
		PollInterval: *poll, ExportInterval: *export,
		Endpoint: *endpoint, CgroupRoot: "/sys/fs/cgroup",
	}
	c := collector.New(cfg, reader, exp)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := c.Start(ctx); err != nil {
		log.Fatalf("start: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("shutting down...")
	cancel()
	c.Stop()
}
