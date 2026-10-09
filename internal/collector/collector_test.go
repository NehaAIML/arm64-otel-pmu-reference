package collector

import (
    "testing"

    "github.com/NehaAIML/arm64-otel-pmu-reference/internal/accum"
)

func TestCollector_PipelineDeterministic(t *testing.T) {
    acc := accum.NewAccumulator()
    key := "0:1234:arm64.pmu.cycles"

    d1 := acc.Update(key, accum.Sample{Counter: 1000, Enabled: 5000, Running: 5000})
    if d1.Value != 0 {
        t.Fatalf("expected initial delta 0, got %d", d1.Value)
    }

    d2 := acc.Update(key, accum.Sample{Counter: 2000, Enabled: 6000, Running: 6000})
    if d2.Value != 1000 {
        t.Errorf("expected arm64.pmu.cycles delta 1000, got %d", d2.Value)
    }
}
