package otlp

import (
	"context"
	"sync"
)

type MetricPoint struct {
	Name       string
	Cumulative float64
	CPU        uint32
	CgroupID   uint64
	Path       string
}
type Exporter struct {
	endpoint string
	testMode bool
	mu       sync.Mutex
	pts      []MetricPoint
}

func NewExporter(endpoint string, testMode bool) *Exporter {
	return &Exporter{endpoint: endpoint, testMode: testMode, pts: make([]MetricPoint, 0)}
}
func (e *Exporter) Start(ctx context.Context) error { return nil }
func (e *Exporter) Export(ctx context.Context, pts []MetricPoint) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.testMode {
		e.pts = append(e.pts, pts...)
	}
	return nil
}
func (e *Exporter) CollectPoints() []MetricPoint { e.mu.Lock(); defer e.mu.Unlock(); return e.pts }
func (e *Exporter) Stop()                        {}
