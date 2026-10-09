package accum

import (
    "testing"
)

func TestAccumulator_Update(t *testing.T) {
    acc := NewAccumulator()
    key := "0:1234:cycles"

    d1 := acc.Update(key, Sample{Counter: 100, Enabled: 1000, Running: 1000})
    if d1.Value != 0 || d1.Reset {
        t.Errorf("expected initial delta 0, no reset; got value=%d, reset=%v", d1.Value, d1.Reset)
    }

    d2 := acc.Update(key, Sample{Counter: 250, Enabled: 2000, Running: 1500})
    if d2.Value != 150 {
        t.Errorf("expected delta value 150, got %d", d2.Value)
    }
    if d2.ScaledValue <= 0 {
        t.Errorf("expected positive scaled value, got %f", d2.ScaledValue)
    }
}

func TestAccumulator_Reset(t *testing.T) {
    acc := NewAccumulator()
    key := "0:1234:cycles"

    acc.Update(key, Sample{Counter: 500, Enabled: 1000, Running: 1000})
    d := acc.Update(key, Sample{Counter: 50, Enabled: 2000, Running: 2000})

    if !d.Reset {
        t.Errorf("expected reset flag to be true")
    }
    if d.Value != 0 {
        t.Errorf("expected value after reset to be 0, got %d", d.Value)
    }
}
