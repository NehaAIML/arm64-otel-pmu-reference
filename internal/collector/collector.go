package collector

import (
	"context"
	"fmt"
	"github.com/NehaAIML/arm64-otel-pmu-reference/internal/accum"
	"github.com/NehaAIML/arm64-otel-pmu-reference/internal/cgroup"
	"github.com/NehaAIML/arm64-otel-pmu-reference/internal/otlp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	PollInterval, ExportInterval time.Duration
	Endpoint, CgroupRoot         string
	SkipCgroupValidation         bool
}
type Snapshot struct { CPU uint32; CgroupID uint64; Cycles, CyclesEn, CyclesRun, Instr, InstrEn, InstrRun, Miss, MissEn, MissRun uint64 }
	CPU                                                                          uint32
	CgroupID                                                                     uint64
	Cycles, CyclesEn, CyclesRun, Instr, InstrEn, InstrRun, Miss, MissEn, MissRun uint64
}
type MapReader interface {
	ForEach(func(Snapshot) error) error
	Delete(Snapshot) error
}
type Collector struct {
	cfg    Config
	acc    *accum.Accumulator
	cg     *cgroup.Manager
	reader MapReader
	exp    *otlp.Exporter
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func New(cfg Config, reader MapReader, exp *otlp.Exporter) *Collector {
	return &Collector{cfg: cfg, acc: accum.NewAccumulator(), cg: cgroup.NewManager(cfg.CgroupRoot), reader: reader, exp: exp, stopCh: make(chan struct{})}
}
func (c *Collector) Start(ctx context.Context) error {
	c.wg.Add(2)
	go c.pollLoop(ctx)
	go c.exportLoop(ctx)
	return nil
}
func (c *Collector) Stop() { close(c.stopCh); c.wg.Wait() }
func (c *Collector) pollLoop(ctx context.Context) {
	defer c.wg.Done()
	tick := time.NewTicker(c.cfg.PollInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-tick.C:
			c.pollOnce()
		}
	}
}
func (c *Collector) pollOnce() {
	live := map[string]bool{}
	validIDs, _ := c.cg.Enumerate()
	isStrict := !c.cfg.SkipCgroupValidation && len(validIDs) > 0
	c.reader.ForEach(func(s Snapshot) error {
		idStr := strconv.FormatUint(s.CgroupID, 10)
		if isStrict {
			if _, ok := validIDs[idStr]; !ok {
				c.reader.Delete(s)
				return nil
			}
		}
		k := fmt.Sprintf("%d/%d/cycles", s.CPU, s.CgroupID)
		c.acc.Update(k, accum.Sample{Counter: s.Cycles, Enabled: s.CyclesEn, Running: s.CyclesRun})
		live[k] = true
		k = fmt.Sprintf("%d/%d/instr", s.CPU, s.CgroupID)
		c.acc.Update(k, accum.Sample{Counter: s.Instr, Enabled: s.InstrEn, Running: s.InstrRun})
		live[k] = true
		k = fmt.Sprintf("%d/%d/miss", s.CPU, s.CgroupID)
		c.acc.Update(k, accum.Sample{Counter: s.Miss, Enabled: s.MissEn, Running: s.MissRun})
		live[k] = true
		return nil
	})
	c.acc.DropStale(live)
}
func (c *Collector) exportLoop(ctx context.Context) {
	defer c.wg.Done()
	tick := time.NewTicker(c.cfg.ExportInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-tick.C:
			c.exportOnce(ctx)
		}
	}
}
func (c *Collector) exportOnce(ctx context.Context) {
	var pts []otlp.MetricPoint
	for _, k := range c.acc.Keys() {
		p := strings.Split(k, "/")
		if len(p) != 3 {
			continue
		}
		cpu, _ := strconv.ParseUint(p[0], 10, 32)
		cg, _ := strconv.ParseUint(p[1], 10, 64)
		pts = append(pts, otlp.MetricPoint{Name: "arm64.pmu." + p[2], Cumulative: c.acc.Scaled(k), CPU: uint32(cpu), CgroupID: cg, Path: c.cg.GetPath(p[1])})
	}
	if len(pts) > 0 {
		c.exp.Export(ctx, pts)
	}
}
