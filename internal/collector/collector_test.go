package collector

import (
	"context"
	"testing"

	"github.com/NehaAIML/arm64-otel-pmu-reference/internal/otlp"
	"github.com/NehaAIML/arm64-otel-pmu-reference/internal/shared"
	"github.com/stretchr/testify/assert"
)

type fakeReader struct {
	snaps []shared.Snapshot
}

func (f *fakeReader) ForEach(fn func(shared.Snapshot) error) error {
	for _, s := range f.snaps {
		if err := fn(s); err != nil {
			return err
		}
	}
	return nil
}
func (f *fakeReader) Delete(s shared.Snapshot) error { return nil }

func TestCollectorPipeline(t *testing.T) {
	fr := &fakeReader{snaps: []shared.Snapshot{{
		CPU: 0, CgroupID: 42,
		Cycles: 1000, CyclesEn: 1000, CyclesRun: 1000,
		Instr: 5000, InstrEn: 1000, InstrRun: 1000,
		Miss: 10, MissEn: 1000, MissRun: 1000,
	}}}
	exp := otlp.NewExporter("localhost:4317", true)
	c := New(Config{SkipCgroupValidation: true}, fr, exp)

	// First poll sets baseline
	c.pollOnce()

	// Second poll generates delta
	fr.snaps[0].Cycles = 2000
	c.pollOnce()

	// Export
	c.exportOnce(context.Background())

	assert.NotEmpty(t, exp.CollectPoints())
}
