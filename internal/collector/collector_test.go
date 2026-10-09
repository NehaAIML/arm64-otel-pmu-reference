package collector

import (
	"context"
	"github.com/NehaAIML/arm64-otel-pmu-reference/internal/otlp"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeReader struct{ snaps []Snapshot }

func (f *fakeReader) ForEach(fn func(Snapshot) error) error {
	for _, s := range f.snaps {
		fn(s)
	}
	return nil
}
func (f *fakeReader) Delete(s Snapshot) error { return nil }
func TestCollectorPipeline(t *testing.T) {
	fr := &fakeReader{snaps: []Snapshot{{CPU: 0, CgroupID: 42, Cycles: 1000, CyclesEn: 1000, CyclesRun: 1000}}}
	exp := otlp.NewExporter("localhost:4317", true)
	c := New(Config{SkipCgroupValidation: true}, fr, exp)
	c.pollOnce()
	fr.snaps[0].Cycles = 2000
	c.pollOnce()
	c.exportOnce(context.Background())
	assert.NotEmpty(t, exp.CollectPoints())
}
